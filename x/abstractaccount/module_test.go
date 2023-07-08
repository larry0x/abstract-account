// there isn't anything to test for in module.go
// instead this file contains helper functions to be used for testing other
// files in this module

package abstractaccount_test

import (
	"encoding/json"
	"errors"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/larry0x/abstract-account/simapp"
	"github.com/larry0x/abstract-account/x/abstractaccount"
	"github.com/larry0x/abstract-account/x/abstractaccount/keeper"
	"github.com/larry0x/abstract-account/x/abstractaccount/testdata"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

const (
	mockChainID = "dev-1"
	signMode    = signing.SignMode_SIGN_MODE_DIRECT
)

func anteTerminator(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	return ctx, nil
}

func postTerminator(ctx sdk.Context, tx sdk.Tx, simulate, success bool) (sdk.Context, error) {
	return ctx, nil
}

func makeBeforeTxDecorator(app *simapp.SimApp) abstractaccount.BeforeTxDecorator {
	return abstractaccount.NewBeforeTxDecorator(app.AbstractAccountKeeper, app.AccountKeeper, app.TxConfig().SignModeHandler())
}

func makeAfterTxDecorator(app *simapp.SimApp) abstractaccount.AfterTxDecorator {
	return abstractaccount.NewAfterTxDecorator(app.AbstractAccountKeeper)
}

func makeMockAccount(keybase keyring.Keyring, uid string) (authtypes.AccountI, error) {
	record, _, err := keybase.NewMnemonic(
		uid,
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	if err != nil {
		return nil, err
	}

	pk, err := record.GetPubKey()
	if err != nil {
		return nil, err
	}

	return authtypes.NewBaseAccount(pk.Address().Bytes(), pk, 0, 0), nil
}

type Signer struct {
	keyName        string             // the name of the key in the keyring
	acc            authtypes.AccountI // the account corresponding to the address
	overrideAccNum *uint64            // if not nil, will override the account number in the AccountI
	overrideSeq    *uint64            // if not nil, will override the sequence in the AccountI
}

func (s *Signer) AccountNumber() uint64 {
	if s.overrideAccNum != nil {
		return *s.overrideAccNum
	}

	return s.acc.GetAccountNumber()
}

func (s *Signer) Sequence() uint64 {
	if s.overrideSeq != nil {
		return *s.overrideSeq
	}

	return s.acc.GetSequence()
}

// Logics in this function is mostly copied from:
// cosmos/cosmos-sdk/x/auth/ante/testutil_test.go/CreateTestTx
func prepareTx(
	ctx sdk.Context, app *simapp.SimApp, keybase keyring.Keyring,
	msgs []sdk.Msg, signers []Signer, chainID string,
	sign bool,
) (authsigning.Tx, error) {
	txBuilder := app.TxConfig().NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// if the tx doesn't need to be signed, we can return here
	if !sign {
		return txBuilder.GetTx(), nil
	}

	// round 1: set empty signature
	sigs := []signing.SignatureV2{}

	for _, signer := range signers {
		sig := signing.SignatureV2{
			PubKey: signer.acc.GetPubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil, // empty
			},
			Sequence: signer.acc.GetSequence(),
		}

		sigs = append(sigs, sig)
	}

	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	// round 2: sign the tx
	sigs = []signing.SignatureV2{}

	for _, signer := range signers {
		signerData := authsigning.SignerData{
			Address:       signer.acc.GetAddress().String(),
			ChainID:       chainID,
			AccountNumber: signer.AccountNumber(),
			Sequence:      signer.Sequence(),
			PubKey:        signer.acc.GetPubKey(),
		}

		signBytes, err := app.TxConfig().SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
		if err != nil {
			return nil, err
		}

		sigBytes, _, err := keybase.Sign(signer.keyName, signBytes)
		if err != nil {
			return nil, err
		}

		sig := signing.SignatureV2{
			PubKey: signer.acc.GetPubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: sigBytes,
			},
			Sequence: signer.Sequence(),
		}

		sigs = append(sigs, sig)
	}

	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

func storeCodeAndRegisterAccount(
	ctx sdk.Context, app *simapp.SimApp, senderAddr sdk.AccAddress,
	bytecode []byte, msg interface{}, funds sdk.Coins,
) (*types.AbstractAccount, error) {
	k := app.AbstractAccountKeeper
	msgServer := keeper.NewMsgServerImpl(k)

	// store code
	codeID, _, err := k.ContractKeeper().Create(ctx, senderAddr, testdata.AccountWasm, nil)
	if err != nil {
		return nil, err
	}

	// prepare the contract instantiate msg
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// register the account
	res, err := msgServer.RegisterAccount(ctx, &types.MsgRegisterAccount{
		Sender: senderAddr.String(),
		CodeID: codeID,
		Msg:    msgBytes,
		Funds:  funds,
		Salt:   []byte("henlo"),
	})
	if err != nil {
		return nil, err
	}

	contractAddr, err := sdk.AccAddressFromBech32(res.Address)
	if err != nil {
		return nil, err
	}

	acc := app.AccountKeeper.GetAccount(ctx, contractAddr)
	if acc == nil {
		return nil, errors.New("account not found")
	}

	abcAcc, ok := acc.(*types.AbstractAccount)
	if !ok {
		return nil, errors.New("account is not an AbstractAccount")
	}

	return abcAcc, nil
}
