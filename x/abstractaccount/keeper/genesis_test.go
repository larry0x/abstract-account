package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/larry0x/abstract-account/simapp"
	simapptesting "github.com/larry0x/abstract-account/simapp/testing"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

var (
	mockParams        = &types.Params{MaxGasBefore: 88888, MaxGasAfter: 99999}
	mockNextAccountID = uint64(12345)
)

func setupGenesisTest() (sdk.Context, *simapp.SimApp) {
	app := simapptesting.MakeSimpleMockApp()
	ctx := app.NewContext(false, tmproto.Header{})

	gs := types.NewGenesisState(mockNextAccountID, mockParams)
	app.AbstractAccountKeeper.InitGenesis(ctx, gs)

	return ctx, app
}

func TestInitGenesis(t *testing.T) {
	ctx, app := setupGenesisTest()

	params, err := app.AbstractAccountKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, mockParams, params)

	nextAccountID := app.AbstractAccountKeeper.GetNextAccountID(ctx)
	require.Equal(t, mockNextAccountID, nextAccountID)
}

func TestExportGenesis(t *testing.T) {
	ctx, app := setupGenesisTest()

	gs := app.AbstractAccountKeeper.ExportGenesis(ctx)
	require.Equal(t, types.NewGenesisState(mockNextAccountID, mockParams), gs)
}
