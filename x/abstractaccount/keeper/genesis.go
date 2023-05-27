package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) []abci.ValidatorUpdate {
	k.SetNextAccountId(ctx, gs.NextAccountId)

	return []abci.ValidatorUpdate{}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		NextAccountId: k.GetNextAccountId(ctx),
	}
}
