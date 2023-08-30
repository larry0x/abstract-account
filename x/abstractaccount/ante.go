package abstractaccount

import (
	"encoding/json"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/keeper"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

var (
	_ sdk.AnteDecorator = &BeforeTxDecorator{}
	_ sdk.PostDecorator = &AfterTxDecorator{}
)

// -------------------------------- GasComsumer --------------------------------

func SigVerificationGasConsumer(
	meter sdk.GasMeter, sig txsigning.SignatureV2, params authtypes.Params,
) error {
	// If the pubkey is a NilPubKey, for now we do not consume any gas (the
	// contract execution consumes it)
	// Otherwise, we simply delegate to the default consumer
	switch sig.PubKey.(type) {
	case *types.NilPubKey:
		return nil
	default:
		return authante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}

// --------------------------------- BeforeTx ----------------------------------

type BeforeTxDecorator struct {
	aak             keeper.Keeper
	ak              authante.AccountKeeper
	signModeHandler authsigning.SignModeHandler
}

func NewBeforeTxDecorator(aak keeper.Keeper, ak authante.AccountKeeper, signModeHandler authsigning.SignModeHandler) BeforeTxDecorator {
	return BeforeTxDecorator{aak, ak, signModeHandler}
}

func (d BeforeTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// first we need to determine whether the rules of account abstraction should
	// apply to this tx. there are two criteria:
	//
	// - the tx has exactly one signer and one signature
	// - this one signer is an AbstractAccount
	//
	// both criteria must be satisfied for this be to be qualified as an AA tx.
	isAbstractAccountTx, signerAcc, sig, err := IsAbstractAccountTx(ctx, tx, d.ak)
	if err != nil {
		return ctx, err
	}

	// if the tx isn't an AA tx, we simply delegate the ante task to the default
	// SigVerificationDecorator
	if !isAbstractAccountTx {
		svd := authante.NewSigVerificationDecorator(d.ak, d.signModeHandler)
		return svd.AnteHandle(ctx, tx, simulate, next)
	}

	// save the account address to the module store. we will need it in the
	// posthandler
	//
	// TODO: a question is that instead of writing to store, can we just put this
	// in memory instead. in practice however, the address is deleted in the post
	// handler, so it's never actually written to disk, meaning the difference in
	// gas consumption should be really small. still worth investigating tho.
	d.aak.SetSignerAddress(ctx, signerAcc.GetAddress())

	// check account sequence number
	if sig.Sequence != signerAcc.GetSequence() {
		return ctx, sdkerrors.ErrWrongSequence.Wrapf("account sequence mismatch, expected %d, got %d", signerAcc.GetSequence(), sig.Sequence)
	}

	// now that we've determined the tx is an AA tx, let us prepare the SudoMsg
	// that will be used to invoke the account contract. The msg includes:
	//
	// - the messages in the tx, converted to []Any
	// - the sign bytes
	// - the credential
	//
	// firstly let's prepare the messages.
	msgAnys, err := sdkMsgsToAnys(tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	// then let us the prepare the sign bytes and credentiale.
	// logics here are mostly copied over from the SigVerificationDecorator.
	signBytes, sigBytes, err := prepareCredentials(ctx, tx, signerAcc, sig.Data, d.signModeHandler)
	if err != nil {
		return ctx, err
	}

	sudoMsgBytes, err := json.Marshal(&types.AccountSudoMsg{
		BeforeTx: &types.BeforeTx{
			Msgs:    msgAnys,
			TxBytes: signBytes,
			// Note that we call this field "cred_bytes" (credental bytes) instead of
			// signature. There is an important reason for this!
			//
			// For EOAs, the credential used to prove a tx is authenticated is a
			// cryptographic signature. For AbstractAccounts however, this is not
			// necessarily the case. The account contract can be programmed to take
			// any credential, not limited to cryptographic signatures. An example of
			// this can be a zk proof that the sender has undergone certain KYC
			// procedures. Therefore, instead of calling this "signature", we choose a
			// more generalized term: credentials.
			CredBytes: sigBytes,
			Simulate:  simulate,
		},
	})
	if err != nil {
		return ctx, err
	}

	params, err := d.aak.GetParams(ctx)
	if err != nil {
		return ctx, err
	}

	if err := sudoWithGasLimit(ctx, d.aak.ContractKeeper(), signerAcc.GetAddress(), sudoMsgBytes, params.MaxGasBefore); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ---------------------------------- AfterTx ----------------------------------

type AfterTxDecorator struct {
	aak keeper.Keeper
}

func NewAfterTxDecorator(aak keeper.Keeper) AfterTxDecorator {
	return AfterTxDecorator{aak}
}

func (d AfterTxDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	// load the signer address, which we determined during the AnteHandler
	//
	// if not found, it means this tx is simply not an AA tx. we skip
	signerAddr := d.aak.GetSignerAddress(ctx)
	if signerAddr == nil {
		return next(ctx, tx, simulate, success)
	}

	d.aak.DeleteSignerAddress(ctx)

	sudoMsgBytes, err := json.Marshal(&types.AccountSudoMsg{
		AfterTx: &types.AfterTx{
			Simulate: simulate,
			// we don't need to pass the `success` parameter into the contract,
			// because the Posthandler is only executed if the tx is successful, so it
			// should always be true anyways
		},
	})
	if err != nil {
		return ctx, err
	}

	params, err := d.aak.GetParams(ctx)
	if err != nil {
		return ctx, err
	}

	if err := sudoWithGasLimit(ctx, d.aak.ContractKeeper(), signerAddr, sudoMsgBytes, params.MaxGasAfter); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate, success)
}

// ---------------------------------- Helpers ----------------------------------

func IsAbstractAccountTx(ctx sdk.Context, tx sdk.Tx, ak authante.AccountKeeper) (bool, *types.AbstractAccount, *txsigning.SignatureV2, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return false, nil, nil, errors.Wrap(sdkerrors.ErrTxDecode, "tx is not a SigVerifiableTx")
	}

	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return false, nil, nil, err
	}

	signerAddrs := sigTx.GetSigners()
	if len(signerAddrs) != 1 || len(sigs) != 1 {
		return false, nil, nil, nil
	}

	signerAcc, err := authante.GetSignerAcc(ctx, ak, signerAddrs[0])
	if err != nil {
		return false, nil, nil, err
	}

	absAcc, ok := signerAcc.(*types.AbstractAccount)
	if !ok {
		return false, nil, nil, nil
	}

	return true, absAcc, &sigs[0], nil
}

func prepareCredentials(
	ctx sdk.Context, tx sdk.Tx, signerAcc authtypes.AccountI,
	sigData txsigning.SignatureData, handler authsigning.SignModeHandler,
) ([]byte, []byte, error) {
	signerData := authsigning.SignerData{
		Address:       signerAcc.GetAddress().String(),
		ChainID:       ctx.ChainID(),
		AccountNumber: signerAcc.GetAccountNumber(), // should we handle the case that this is a gentx?
		Sequence:      signerAcc.GetSequence(),
		PubKey:        signerAcc.GetPubKey(),
	}

	data, ok := sigData.(*txsigning.SingleSignatureData)
	if !ok {
		return nil, nil, types.ErrNotSingleSignautre
	}

	signBytes, err := handler.GetSignBytes(data.SignMode, signerData, tx)
	if err != nil {
		return nil, nil, err
	}

	return signBytes, data.Signature, nil
}

func sdkMsgsToAnys(msgs []sdk.Msg) ([]*types.Any, error) {
	anys := []*types.Any{}

	for _, msg := range msgs {
		msgAny, err := types.NewAnyFromProtoMsg(msg)
		if err != nil {
			return nil, err
		}

		anys = append(anys, msgAny)
	}

	return anys, nil
}

// Call a contract's sudo entry point with a gas limit.
//
// Copied from Osmosis' protorev posthandler:
// https://github.com/osmosis-labs/osmosis/blob/98025f185ab2ee1b060511ed22679112abcc08fa/x/protorev/keeper/posthandler.go#L42-L43
//
// Thanks Roman and Jorge for the helpful discussion.
func sudoWithGasLimit(
	ctx sdk.Context, contractKeeper wasmtypes.ContractOpsKeeper,
	contractAddr sdk.AccAddress, msg []byte, maxGas sdk.Gas,
) error {
	cacheCtx, write := ctx.CacheContext()
	cacheCtx = cacheCtx.WithGasMeter(sdk.NewGasMeter(maxGas))

	if _, err := contractKeeper.Sudo(cacheCtx, contractAddr, msg); err != nil {
		return err
	}

	write()
	ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

	return nil
}
