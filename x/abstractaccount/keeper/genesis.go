package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) []abci.ValidatorUpdate {
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	k.SetNextAccountID(ctx, gs.NextAccountId)

	return []abci.ValidatorUpdate{}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:        params,
		NextAccountId: k.GetNextAccountID(ctx),
	}
}
