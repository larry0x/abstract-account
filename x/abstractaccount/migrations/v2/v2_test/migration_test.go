package v2_test

import (
	"testing"

	cometdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v2 "github.com/larry0x/abstract-account/x/abstractaccount/migrations/v2"
)

func TestMigrateStore(t *testing.T) {
	db := cometdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)

	storeKey := sdk.NewKVStoreKey(paramtypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_key")

	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	store := ctx.KVStore(storeKey)
	require.NoError(t, stateStore.LoadLatestVersion())

	bz := store.Get(types.KeyParams)
	require.Nil(t, bz)

	// run global fee migration
	err := v2.MigrateStore(ctx, storeKey, cdc)
	require.NoError(t, err)

	// get params
	storeAfterMig := ctx.KVStore(storeKey)
	newBz := storeAfterMig.Get(types.KeyParams)
	require.NotNil(t, newBz)

	var newParams types.Params
	require.NoError(t, cdc.Unmarshal(newBz, &newParams))

	expectedParams := types.DefaultParams()
	require.Equal(t, expectedParams.AllowAllCodeIDs, newParams.AllowAllCodeIDs)
	require.Equal(t, 0, len(newParams.AllowedCodeIDs))
	require.Equal(t, expectedParams.MaxGasBefore, newParams.MaxGasBefore)
	require.Equal(t, expectedParams.MaxGasAfter, newParams.MaxGasAfter)

}
