package abstractaccount_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapptesting "github.com/larry0x/abstract-account/simapp/testing"
	"github.com/larry0x/abstract-account/x/abstractaccount"
	"github.com/larry0x/abstract-account/x/abstractaccount/testdata"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

const (
	mockChainID = "dev-1"
	mockAccNum  = uint64(12345)
	mockSeq     = uint64(88888)
	signMode    = signing.SignMode_SIGN_MODE_DIRECT
)

func TestIsAbstractAccountTx(t *testing.T) {
	var (
		app     = simapptesting.MakeSimpleMockApp()
		ctx     = app.NewContext(false, tmproto.Header{})
		keybase = keyring.NewInMemory(app.Codec())
	)

	// we create two mock accounts: 1, a BaseAccount, 2, an AbstractAccount
	acc1, err := makeMockAccount(keybase, "test1")
	require.NoError(t, err)

	acc2, err := makeMockAccount(keybase, "test2")
	require.NoError(t, err)

	app.AccountKeeper.SetAccount(ctx, acc1)
	app.AccountKeeper.SetAccount(ctx, types.NewAbstractAccountFromAccount(acc2))

	for _, tc := range []struct {
		desc  string
		msgs  []sdk.Msg
		expIs bool
	}{
		{
			desc: "tx has one signer and it is an AbstractAccount",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
			},
			expIs: true,
		},
		{
			desc: "tx has one signer but it's not an AbstractAccount",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
			},
			expIs: false,
		},
		{
			desc: "tx has more than one signers",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
			},
			expIs: false,
		},
	} {
		sigTx, err := prepareTx(ctx, app, tc.msgs)
		require.NoError(t, err)

		is, _, _, err := abstractaccount.IsAbstractAccountTx(ctx, sigTx, app.AccountKeeper)
		require.NoError(t, err)
		require.Equal(t, tc.expIs, is)
	}
}

func TestBeforeTx(t *testing.T) {
	var (
		app     = simapptesting.MakeSimpleMockApp()
		keybase = keyring.NewInMemory(app.Codec())
	)

	ctx := app.NewContext(false, tmproto.Header{
		// whenever we execute a contract, we must specify the block time in the
		// header, so that wasmkeeper knows what to use for env.block.time
		//
		// if not doing this, will get this error:
		// panic: Block (unix) time must never be empty or negative
		Time: time.Now(),

		// whenever we want to do signature verification of a tx, we must specify
		// the chain-id in the header, otherwise it defaults to an empty string and
		// verification will always fail
		ChainID: mockChainID,
	})

	// create two mock accounts
	acc1, err := makeMockAccount(keybase, "test1")
	require.NoError(t, err)

	acc2, err := makeMockAccount(keybase, "test2")
	require.NoError(t, err)

	// register the AbstractAccount
	absAcc, err := storeCodeAndRegisterAccount(
		ctx,
		app,
		// use the pubkey of acc1 as the AbstractAccount's pubkey
		acc1.GetAddress(),
		testdata.AccountWasm,
		&AccountInitMsg{PubKey: acc1.GetPubKey().Bytes()},
		sdk.NewCoins(),
	)
	require.NoError(t, err)

	// change the AbstractAccount's account number and sequence to some non-zero
	// numbers to make the tests harder
	app.AccountKeeper.RemoveAccount(ctx, absAcc)
	absAcc.SetAccountNumber(mockAccNum)
	absAcc.SetSequence(mockSeq)
	app.AccountKeeper.SetAccount(ctx, absAcc)

	for _, tc := range []struct {
		desc     string
		signWith string
		chainID  string
		accNum   uint64
		seq      uint64
		expOk    bool
	}{
		{
			desc:     "tx signed with the correct key",
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			expOk:    true,
		},
		{
			desc:     "tx signed with an incorrect key",
			signWith: "test2",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			expOk:    false,
		},
		{
			desc:     "tx signed with an incorrect chain id",
			signWith: "test1",
			chainID:  "wrong-chain-id",
			accNum:   mockAccNum,
			seq:      mockSeq,
			expOk:    false,
		},
		{
			desc:     "tx signed with an incorrect account number",
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   4524455,
			seq:      mockSeq,
			expOk:    false,
		},
		{
			desc:     "tx signed with an incorrect sequence",
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      5786786,
			expOk:    false,
		},
	} {
		msg := banktypes.NewMsgSend(absAcc.GetAddress(), acc2.GetAddress(), sdk.NewCoins())

		tx, err := prepareTx2(
			ctx,
			app,
			[]sdk.Msg{msg},
			keybase,
			tc.signWith,
			absAcc,
			tc.chainID,
			tc.accNum,
			tc.seq,
		)
		require.NoError(t, err)

		decorator := makeBeforeTxDecorator(app)
		_, err = decorator.AnteHandle(ctx, tx, false, anteTerminator)

		if tc.expOk {
			require.NoError(t, err)

			// the signer address should have been stored for use by the PostHandler
			signerAddr := app.AbstractAccountKeeper.GetSignerAddress(ctx)
			require.Equal(t, absAcc.GetAddress(), signerAddr)

			// delete the stored signer address so that we start from a clean state
			// for the next test case
			app.AbstractAccountKeeper.DeleteSignerAddress(ctx)
		} else {
			require.Error(t, err)
		}
	}
}

func TestAfterTx(t *testing.T) {
	var (
		app     = simapptesting.MakeSimpleMockApp()
		keybase = keyring.NewInMemory(app.Codec())
	)

	ctx := app.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		ChainID: mockChainID,
	})

	// create a mock account
	acc, err := makeMockAccount(keybase, "test1")
	require.NoError(t, err)

	// register the AbstractAccount
	absAcc, err := storeCodeAndRegisterAccount(
		ctx,
		app,
		acc.GetAddress(),
		testdata.AccountWasm,
		&AccountInitMsg{PubKey: acc.GetPubKey().Bytes()},
		sdk.NewCoins(),
	)
	require.NoError(t, err)

	// save the signer address to mimic what happens in the BeforeTx hook
	app.AbstractAccountKeeper.SetSignerAddress(ctx, absAcc.GetAddress())

	tx, err := prepareTx2(
		ctx,
		app,
		[]sdk.Msg{banktypes.NewMsgSend(absAcc.GetAddress(), acc.GetAddress(), sdk.NewCoins())},
		keybase,
		"test1",
		absAcc,
		mockChainID,
		absAcc.GetAccountNumber(),
		absAcc.GetSequence(),
	)
	require.NoError(t, err)

	decorator := makeAfterTxDecorator(app)
	_, err = decorator.PostHandle(ctx, tx, false, true, postTerminator)
	require.NoError(t, err)
}
