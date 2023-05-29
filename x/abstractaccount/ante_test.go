package abstractaccount_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapptesting "github.com/larry0x/abstract-account/simapp/testing"
	"github.com/larry0x/abstract-account/x/abstractaccount"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func TestIsAbstractAccountTx(t *testing.T) {
	var (
		app     = simapptesting.MakeSimpleMockApp()
		ctx     = app.NewContext(false, tmproto.Header{})
		keybase = keyring.NewInMemory(app.Codec())
	)

	// we create two mock accounts: 1 a BaseAccount, 2 an AbstractAccount
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
			desc: "tx has more than one signers",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
			},
			expIs: false,
		},
		{
			desc: "tx has one signer but it's not an AbstractAccount",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
				banktypes.NewMsgSend(acc1.GetAddress(), acc2.GetAddress(), sdk.NewCoins()),
			},
			expIs: false,
		},
		{
			desc: "tx has one signer and it is an AbstractAccount",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
				banktypes.NewMsgSend(acc2.GetAddress(), acc1.GetAddress(), sdk.NewCoins()),
			},
			expIs: true,
		},
	} {
		sigTx, err := prepareTx(ctx, app, keybase, tc.msgs)
		require.NoError(t, err)

		is, _, _, err := abstractaccount.IsAbstractAccountTx(ctx, sigTx, app.AccountKeeper)
		require.NoError(t, err)
		require.Equal(t, tc.expIs, is)
	}
}

func TestBeforeTx(t *testing.T) {
	// TODO
}

func TestAfterTx(t *testing.T) {
	// TODO
}
