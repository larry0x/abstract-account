// there isn't anything to test for in module.go
// instead this file contains helper functions to be used for testing ante.go

package abstractaccount_test

import (
	"encoding/json"
	"errors"
	"fmt"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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

const mockChainID = "dev-1"

func makeBeforeTxDecoratorForTesting(app *simapp.SimApp) abstractaccount.BeforeTxDecorator {
	return abstractaccount.NewBeforeTxDecorator(
		app.AbstractAccountKeeper,
		app.AccountKeeper,
		app.TxConfig().SignModeHandler(),
	)
}

func makeAfterTxDecoratorForTesting(app *simapp.SimApp) abstractaccount.AfterTxDecorator {
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
func prepareTx(ctx sdk.Context, app *simapp.SimApp, keybase keyring.Keyring, msgs []sdk.Msg) (authsigning.Tx, error) {
	signMode := app.TxConfig().SignModeHandler().DefaultMode()

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

	// round 1: set empty signatures
	sigs := []signing.SignatureV2{}
	for _, signerAcc := range signerAccs {
		sigData := &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		}

		sig := signing.SignatureV2{
			PubKey:   signerAcc.GetPubKey(),
			Data:     sigData,
			Sequence: signerAcc.GetSequence(),
		}

		sigs = append(sigs, sig)
	}

	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return nil, err
	}

	// round 2: all signer infos are set, each signer can sign
	sigs = []signing.SignatureV2{}
	for _, signerAcc := range signerAccs {
		r, err := keybase.KeyByAddress(signerAcc.GetAddress())
		if err != nil {
			return nil, err
		}

		rl := r.GetLocal()
		if rl == nil {
			return nil, keyring.ErrPrivKeyExtr
		}

		if rl.PrivKey == nil {
			return nil, errors.New("private key is not available")
		}

		privKey, ok := rl.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
		if !ok {
			return nil, errors.New("unable to cast any to cryptotypes.PrivKey")
		}

		signerData := authsigning.SignerData{
			Address:       signerAcc.GetAddress().String(),
			ChainID:       mockChainID,
			AccountNumber: signerAcc.GetAccountNumber(),
			Sequence:      signerAcc.GetSequence(),
			PubKey:        signerAcc.GetPubKey(),
		}

		sig, err := clienttx.SignWithPrivKey(
			signMode,
			signerData,
			txBuilder,
			privKey,
			app.TxConfig(),
			signerAcc.GetSequence(),
		)
		if err != nil {
			return nil, err
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
) (codeID uint64, contractAddr sdk.Address, err error) {
	k := app.AbstractAccountKeeper
	msgServer := keeper.NewMsgServerImpl(k)

	// store code
	codeID, _, err = k.ContractKeeper().Create(ctx, senderAddr, testdata.AccountWasm, nil)
	if err != nil {
		return 0, nil, err
	}

	// prepare the contract instantiate msg
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return 0, nil, err
	}

	// register the account
	res, err := msgServer.RegisterAccount(ctx, &types.MsgRegisterAccount{
		Sender: senderAddr.String(),
		CodeID: codeID,
		Msg:    msgBytes,
		Funds:  funds,
	})
	if err != nil {
		return 0, nil, err
	}

	contractAddr, err = sdk.AccAddressFromBech32(res.Address)
	if err != nil {
		return 0, nil, err
	}

	return codeID, contractAddr, nil
}
