package keeper

import (
	log "github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	accountKeeper authkeeper.AccountKeeperI
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, accountKeeper authkeeper.AccountKeeperI) Keeper {
	return Keeper{cdc, storeKey, accountKeeper}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetSignerAddress(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	return store.Get(types.KeySignerAddress)
}

func (k Keeper) SetSignerAddress(ctx sdk.Context, signerAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeySignerAddress, signerAddr)
}

func (k Keeper) DeleteSignerAddress(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeySignerAddress)
}
