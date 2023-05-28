package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	simapptesting "github.com/larry0x/abstract-account/simapp/testing"
)

func TestGetAndIncrementNextAccountID(t *testing.T) {
	app := simapptesting.MakeSimpleMockApp()
	ctx := app.NewContext(false, tmproto.Header{})

	id := app.AbstractAccountKeeper.GetAndIncrementNextAccountID(ctx)
	require.Equal(t, uint64(1), id)

	id = app.AbstractAccountKeeper.GetNextAccountID(ctx)
	require.Equal(t, uint64(2), id)
}
