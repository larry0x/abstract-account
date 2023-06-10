package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/larry0x/abstract-account/simapp"
	simapptesting "github.com/larry0x/abstract-account/simapp/testing"
	"github.com/larry0x/abstract-account/x/abstractaccount/keeper"
	"github.com/larry0x/abstract-account/x/abstractaccount/testdata"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

type AccountInitMsg struct {
	PubKey []byte `json:"pubkey"`
}

var (
	user               = simapptesting.MakeRandomAddress()
	userInitialBalance = sdk.NewCoins(sdk.NewCoin(simapptesting.DefaultBondDenom, sdk.NewInt(123456)))
	acctRegisterFunds  = sdk.NewCoins(sdk.NewCoin(simapptesting.DefaultBondDenom, sdk.NewInt(88888)))
)

func TestUpdateParams(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		sender    string
		newParams *types.Params
		expErr    bool
	}{
		{
			desc:      "sender is not authority",
			sender:    user.String(),
			newParams: types.DefaultParams(),
			expErr:    true,
		},
		{
			desc:      "invalid params",
			sender:    simapp.Authority,
			newParams: &types.Params{MaxGasBefore: 88888, MaxGasAfter: 0},
			expErr:    true,
		},
		{
			desc:      "sender is authority and params are valid",
			sender:    simapp.Authority,
			newParams: &types.Params{MaxGasBefore: 88888, MaxGasAfter: 99999},
			expErr:    false,
		},
	} {
		app := simapptesting.MakeMockApp([]banktypes.Balance{})
		ctx := app.NewContext(false, tmproto.Header{})

		msgServer := keeper.NewMsgServerImpl(app.AbstractAccountKeeper)

		paramsBefore, err1 := app.AbstractAccountKeeper.GetParams(ctx)
		require.NoError(t, err1)

		_, err2 := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Sender: tc.sender,
			Params: tc.newParams,
		})

		paramsAfter, err3 := app.AbstractAccountKeeper.GetParams(ctx)
		require.NoError(t, err3)

		if tc.expErr {
			require.Error(t, err2)
			require.Equal(t, paramsBefore, paramsAfter)
		} else {
			require.NoError(t, err2)
			require.Equal(t, tc.newParams, paramsAfter)
		}
	}
}

func TestRegisterAccount(t *testing.T) {
	app := simapptesting.MakeMockApp([]banktypes.Balance{
		{
			Address: user.String(),
			Coins:   userInitialBalance,
		},
	})

	ctx := app.NewContext(false, tmproto.Header{
		// whenever we execute a contract, we must specify the block time in the
		// header, so that wasmkeeper knows what to use for env.block.time
		//
		// if not doing this, will get this error:
		// panic: Block (unix) time must never be empty or negative
		Time: time.Now(),
	})

	k := app.AbstractAccountKeeper
	msgServer := keeper.NewMsgServerImpl(k)

	// store code
	codeID, _, err := k.ContractKeeper().Create(ctx, user, testdata.AccountWasm, nil)
	require.NoError(t, err)

	// prepare the contract instantiate msg
	msgBytes, err := json.Marshal(&AccountInitMsg{
		PubKey: simapptesting.MakeRandomPubKey().Bytes(),
	})
	require.NoError(t, err)

	// register the account
	res, err := msgServer.RegisterAccount(ctx, &types.MsgRegisterAccount{
		Sender: user.String(),
		CodeID: codeID,
		Msg:    msgBytes,
		Funds:  acctRegisterFunds,
		Salt:   []byte("hello"),
	})
	require.NoError(t, err)

	contractAddr, err := sdk.AccAddressFromBech32(res.Address)
	require.NoError(t, err)

	// check the contract info is correct
	contractInfo := app.WasmKeeper.GetContractInfo(ctx, contractAddr)
	require.Equal(t, codeID, contractInfo.CodeID)
	require.Equal(t, user.String(), contractInfo.Creator)
	require.Equal(t, contractAddr.String(), contractInfo.Admin)
	require.Equal(t, fmt.Sprintf("%s/%d", types.ModuleName, k.GetNextAccountID(ctx)-1), contractInfo.Label)

	// make sure an AbstractAccount has been created
	_, ok := app.AccountKeeper.GetAccount(ctx, contractAddr).(*types.AbstractAccount)
	require.True(t, ok)

	// make sure the contract has received the funds
	balance := app.BankKeeper.GetAllBalances(ctx, contractAddr)
	require.Equal(t, acctRegisterFunds, balance)
}
