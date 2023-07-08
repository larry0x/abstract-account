package abstractaccount_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapptesting "github.com/larry0x/abstract-account/simapp/testing"
	"github.com/larry0x/abstract-account/x/abstractaccount"
	"github.com/larry0x/abstract-account/x/abstractaccount/testdata"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
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
	acc2 = types.NewAbstractAccountFromAccount(acc2)
	require.NoError(t, err)

	app.AccountKeeper.SetAccount(ctx, acc1)
	app.AccountKeeper.SetAccount(ctx, acc2)

	signer1 := Signer{
		keyName:        "test1",
		acc:            acc1,
		overrideAccNum: nil,
		overrideSeq:    nil,
	}
	signer2 := Signer{
		keyName:        "test2",
		acc:            acc2,
		overrideAccNum: nil,
		overrideSeq:    nil,
	}

	for _, tc := range []struct {
		desc    string
		msgs    []sdk.Msg
		signers []Signer
		expIs   bool
	}{
		{
			desc: "tx has one signer and it is an AbstractAccount",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
			},
			signers: []Signer{signer2},
			expIs:   true,
		},
		{
			desc: "tx has one signer but it's not an AbstractAccount",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
			},
			signers: []Signer{signer1},
			expIs:   false,
		},
		{
			desc: "tx has more than one signers",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
			},
			signers: []Signer{signer1, signer2},
			expIs:   false,
		},
	} {
		sigTx, err := prepareTx(ctx, app, keybase, tc.msgs, tc.signers, mockChainID, true)
		require.NoError(t, err)

		is, _, _, err := abstractaccount.IsAbstractAccountTx(ctx, sigTx, app.AccountKeeper)
		require.NoError(t, err)
		require.Equal(t, tc.expIs, is)
	}
}

type BaseInstantiateMsg struct {
	PubKey []byte `json:"pubkey"`
}

func TestBeforeTx(t *testing.T) {
	var (
		app        = simapptesting.MakeSimpleMockApp()
		keybase    = keyring.NewInMemory(app.Codec())
		mockAccNum = uint64(12345)
		mockSeq    = uint64(88888)
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
		&BaseInstantiateMsg{PubKey: acc1.GetPubKey().Bytes()},
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
		simulate bool   // whether to run the AnteHandler in simulation mode
		sign     bool   // whether a signature is to be included with this tx
		signWith string // if a sig is to be included, which key to use to sign it
		chainID  string
		accNum   uint64
		seq      uint64
		maxGas   uint64
		expOk    bool
		expPanic bool
	}{
		{
			desc:     "tx signed with the correct key",
			simulate: false,
			sign:     true,
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    true,
			expPanic: false,
		},
		{
			desc:     "tx signed with an incorrect key",
			simulate: false,
			sign:     true,
			signWith: "test2",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    false,
			expPanic: false,
		},
		{
			desc:     "tx signed with an incorrect chain id",
			simulate: false,
			sign:     true,
			signWith: "test1",
			chainID:  "wrong-chain-id",
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    false,
			expPanic: false,
		},
		{
			desc:     "tx signed with an incorrect account number",
			simulate: false,
			sign:     true,
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   4524455,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    false,
			expPanic: false,
		},
		{
			desc:     "tx signed with an incorrect sequence",
			simulate: false,
			sign:     true,
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      5786786,
			maxGas:   types.DefaultMaxGas,
			expOk:    false,
			expPanic: false,
		},
		{
			desc:     "contract call exceeds gas limit",
			simulate: false,
			sign:     true,
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   1, // the call for sure will use more than 1 gas
			expOk:    false,
			expPanic: true, // attempting to consume above the gas limit results in panicking
		},
		{
			desc:     "not in simulation mode, but tx isn't signed",
			simulate: false,
			sign:     false,
			signWith: "",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    false,
			expPanic: false,
		},
		{
			desc:     "in simulation, tx is signed",
			simulate: true,
			sign:     true,
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    true, // we accept it
			expPanic: false,
		},
		{
			desc:     "in simulation, tx is not signed",
			simulate: true,
			sign:     false,
			signWith: "test1",
			chainID:  mockChainID,
			accNum:   mockAccNum,
			seq:      mockSeq,
			maxGas:   types.DefaultMaxGas,
			expOk:    true, // in simulation mode, for this particular account type, the credential can be omitted
			expPanic: false,
		},
	} {
		// set max gas
		app.AbstractAccountKeeper.SetParams(ctx, &types.Params{
			MaxGasBefore: tc.maxGas,
			MaxGasAfter:  types.DefaultMaxGas,
		})

		msg := banktypes.NewMsgSend(absAcc.GetAddress(), acc2.GetAddress(), sdk.NewCoins())

		signer := Signer{
			keyName:        tc.signWith,
			acc:            absAcc,
			overrideAccNum: &tc.accNum,
			overrideSeq:    &tc.seq,
		}

		tx, err := prepareTx(
			ctx,
			app,
			keybase,
			[]sdk.Msg{msg},
			[]Signer{signer},
			tc.chainID,
			tc.sign,
		)
		require.NoError(t, err)

		if tc.expPanic {
			require.Panics(t, func() {
				decorator := makeBeforeTxDecorator(app)
				decorator.AnteHandle(ctx, tx, tc.simulate, anteTerminator)
			})

			return
		}

		decorator := makeBeforeTxDecorator(app)
		_, err = decorator.AnteHandle(ctx, tx, tc.simulate, anteTerminator)

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
		&BaseInstantiateMsg{PubKey: acc.GetPubKey().Bytes()},
		sdk.NewCoins(),
	)
	require.NoError(t, err)

	// save the signer address to mimic what happens in the BeforeTx hook
	app.AbstractAccountKeeper.SetSignerAddress(ctx, absAcc.GetAddress())

	tx, err := prepareTx(
		ctx,
		app,
		keybase,
		[]sdk.Msg{banktypes.NewMsgSend(absAcc.GetAddress(), acc.GetAddress(), sdk.NewCoins())},
		[]Signer{{
			keyName: "test1",
			acc:     absAcc,
		}},
		mockChainID,
		true,
	)
	require.NoError(t, err)

	decorator := makeAfterTxDecorator(app)
	_, err = decorator.PostHandle(ctx, tx, false, true, postTerminator)
	require.NoError(t, err)
}
