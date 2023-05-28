package keeper

import (
	log "github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	ak       authkeeper.AccountKeeperI
	ck       wasmtypes.ContractOpsKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec, storeKey storetypes.StoreKey,
	ak authkeeper.AccountKeeperI, ck wasmtypes.ContractOpsKeeper,
) Keeper {
	if ak == nil {
		panic("AccountKeeperI cannot be nil")
	}

	if ck == nil {
		panic("ContractOpsKeeper cannot be nil")
	}

	return Keeper{cdc, storeKey, ak, ck}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) ContractKeeper() wasmtypes.ContractOpsKeeper {
	return k.ck
}

// ------------------------------- NextAccountId -------------------------------

func (k Keeper) GetAndIncrementNextAccountID(ctx sdk.Context) uint64 {
	id := k.GetNextAccountID(ctx)

	k.SetNextAccountID(ctx, id+1)

	return id
}

func (k Keeper) GetNextAccountID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	return sdk.BigEndianToUint64(store.Get(types.KeyNextAccountID))
}

func (k Keeper) SetNextAccountID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyNextAccountID, sdk.Uint64ToBigEndian(id))
}

// ------------------------------- SignerAddress -------------------------------

func (k Keeper) GetSignerAddress(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)

	return sdk.AccAddress(store.Get(types.KeySignerAddress))
}

func (k Keeper) SetSignerAddress(ctx sdk.Context, signerAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeySignerAddress, signerAddr)
}

func (k Keeper) DeleteSignerAddress(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeySignerAddress)
}
