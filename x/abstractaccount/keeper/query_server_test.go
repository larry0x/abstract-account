package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/larry0x/abstract-account/x/abstractaccount/keeper"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func TestQueryParams(t *testing.T) {
	ctx, app := setupGenesisTest()

	queryServer := keeper.NewQueryServerImpl(app.AbstractAccountKeeper)

	res, err := queryServer.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, mockParams, res.Params)
}
