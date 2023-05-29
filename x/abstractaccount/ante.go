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
	// First we need to determine whether the rules of account abstraction should
	// apply to this tx. There are two criteria:
	//
	// - The tx has exactly one signer and one signature
	// - This one signer is an AbstractAccount
	//
	// Both criteria must be satisfied for this be to be qualified as an AA tx.
	isAbstractAccountTx, signerAcc, sig, err := IsAbstractAccountTx(ctx, tx, d.ak)
	if err != nil {
		return ctx, err
	}

	// If the tx is an AA tx, we save the signer address to the module store.
	// We will use it in the PostHandler.
	//
	// If it's not an AA tx, we simply delegate the ante task to the default
	// SigVerificationDecorator.
	if isAbstractAccountTx {
		d.aak.SetSignerAddress(ctx, signerAcc.GetAddress())
	} else {
		svd := authante.NewSigVerificationDecorator(d.ak, d.signModeHandler)
		return svd.AnteHandle(ctx, tx, simulate, next)
	}

	// Check account sequence number
	if sig.Sequence != signerAcc.GetSequence() {
		return ctx, sdkerrors.ErrWrongSequence.Wrapf("account sequence mismatch, expected %d, got %d", signerAcc.GetSequence(), sig.Sequence)
	}

	// Now that we've determined the tx is an AA tx, let us prepare the SudoMsg
	// that will be used to invoke the account contract. The msg includes:
	//
	// - The messages in the tx, converted to []Any
	// - The sign bytes
	// - The credential
	//
	// Firstly let's prepare the messages.
	msgAnys, err := sdkMsgsToAnys(tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	// Then let us the prepare the sign bytes and credentiale.
	// Logics here are mostly copied over from the SigVerificationDecorator.
	signBytes, sigBytes, err := prepareCredentials(ctx, tx, signerAcc, sig.Data, d.signModeHandler)
	if err != nil {
		return ctx, err
	}

	sudoMsgBytes, err := json.Marshal(&types.AccountSudoMsg{
		BeforeTx: &types.BeforeTx{
			Msgs:    msgAnys,
			TxBytes: signBytes,
			// Note that we call this field "credential" instead of signature. There
			// is an important reason for this!
			//
			// For EOAs, the credential used to prove a tx is authenticated is a
			// cryptographic signature. For AbstractAccounts however, this is not
			// necessarily the case. The account contract can be programmed to take
			// any credential, not limited to cryptographic signatures. An example of
			// this can be a zk proof that the sender has undergone certain KYC
			// procedures. Therefore, instead of calling this "signature", we choose a
			// more generalized term: credentials.
			Credential: sigBytes,
		},
	})
	if err != nil {
		return ctx, err
	}

	_, err = d.aak.ContractKeeper().Sudo(ctx, signerAcc.GetAddress(), sudoMsgBytes)
	if err != nil {
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
	// Load the signer address, which we determined during the AnteHandler.
	//
	// If found, we delete it from module store since it's not needed for handling
	// the next tx.
	//
	// If not found, it means this tx is not an AA tx, in which case we skip.
	signerAddr := d.aak.GetSignerAddress(ctx)
	if signerAddr != nil {
		d.aak.DeleteSignerAddress(ctx)
	} else {
		return next(ctx, tx, simulate, success)
	}

	sudoMsgBytes, err := json.Marshal(&types.AccountSudoMsg{
		AfterTx: &types.AfterTx{
			Success: success,
		},
	})
	if err != nil {
		return ctx, err
	}

	_, err = d.aak.ContractKeeper().Sudo(ctx, signerAddr, sudoMsgBytes)
	if err != nil {
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
