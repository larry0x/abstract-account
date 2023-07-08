// there isn't anything to test for in module.go
// instead this file contains helper functions to be used for testing other
// files in this module

package abstractaccount_test

import (
	"encoding/json"
	"errors"
	"fmt"

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

type AccountInitMsg struct {
	PubKey []byte `json:"pubkey"`
}

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

// Logics in this function is mostly copied from:
// cosmos/cosmos-sdk/x/auth/ante/testutil_test.go/CreateTestTx
func prepareTx(ctx sdk.Context, app *simapp.SimApp, msgs []sdk.Msg) (authsigning.Tx, error) {
	txBuilder := app.TxConfig().NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// gather signer accounts
	signerAccs := []authtypes.AccountI{}
	for _, signerAddr := range txBuilder.GetTx().GetSigners() {
		signerAcc := app.AccountKeeper.GetAccount(ctx, signerAddr)
		if signerAcc == nil {
			return nil, fmt.Errorf("account not found for signer %s", signerAddr.String())
		}

		signerAccs = append(signerAccs, signerAcc)
	}

	// set an empty signature for each signer
	// for this test we don't need valid signatures
	sigs := []signing.SignatureV2{}
	for _, signerAcc := range signerAccs {
		sig := signing.SignatureV2{
			PubKey: signerAcc.GetPubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil,
			},
			Sequence: signerAcc.GetSequence(),
		}

		sigs = append(sigs, sig)
	}

	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

// Similar to prepareTx, this function also takes a slice of messages, compose
// a tx, and sign it. The differences are that
//  1. this function only works for txs that has exactly one signer, and this
//     signer is an AbstractAccount
//  2. this function allows altering the signer account, chain-id, sequence, and
//     account-number
func prepareTx2(
	ctx sdk.Context, app *simapp.SimApp, msgs []sdk.Msg,
	keybase keyring.Keyring, keyName string, sign bool,
	absAcc *types.AbstractAccount, chainID string, accNum, seq uint64,
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
	sig := signing.SignatureV2{
		PubKey: absAcc.GetPubKey(), // NilPubKey
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil, // empty
		},
		Sequence: seq,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	// round 2: sign the tx
	signerData := authsigning.SignerData{
		Address:       absAcc.GetAddress().String(),
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      seq,
		PubKey:        absAcc.GetPubKey(), // NilPubKey
	}

	signBytes, err := app.TxConfig().SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	sigBytes, _, err := keybase.Sign(keyName, signBytes)
	if err != nil {
		return nil, err
	}

	sig = signing.SignatureV2{
		PubKey: absAcc.GetPubKey(), // NilPubKey
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: sigBytes,
		},
		Sequence: seq,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
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
