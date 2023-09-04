package keeper

import (
	log "github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	ak        authkeeper.AccountKeeperI
	ck        wasmtypes.ContractOpsKeeper
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec, storeKey storetypes.StoreKey,
	ak authkeeper.AccountKeeperI, ck wasmtypes.ContractOpsKeeper,
	authority string,
) Keeper {
	if ak == nil {
		panic("AccountKeeperI cannot be nil")
	}

	if ck == nil {
		panic("ContractOpsKeeper cannot be nil")
	}

	return Keeper{cdc, storeKey, ak, ck, authority}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) ContractKeeper() wasmtypes.ContractOpsKeeper {
	return k.ck
}

// ---------------------------------- Params -----------------------------------

func (k Keeper) GetParams(ctx sdk.Context) (*types.Params, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyParams)
	if bz == nil {
		return nil, sdkerrors.ErrNotFound.Wrap("x/abstractaccount module params")
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, types.ErrParsingParams.Wrap(err.Error())
	}

	return &params, nil
}

func (k Keeper) SetParams(ctx sdk.Context, params *types.Params) error {
	store := ctx.KVStore(k.storeKey)

	// params must be valid before we save it
	// there are two instances where SetParams is called - in Keeper.InitGenesis,
	// and in msgServer.UpdateParams
	// we can either perform the validation in those two functions, or do it
	// together here. doing it here seems cleaner.
	if err := params.Validate(); err != nil {
		return err
	}

	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return types.ErrParsingParams.Wrap(err.Error())
	}

	store.Set(types.KeyParams, bz)

	return nil
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

// ------------------------------- Migration -------------------------------
func (k Keeper) Migrator() Migrator {
	return NewMigrator(k.storeKey, k.cdc)
}
