package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/larry0x/abstract-account/x/abstractaccount/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	key storetypes.StoreKey
	cdc codec.BinaryCodec
}

// NewMigrator returns a new Migrator.
func NewMigrator(key storetypes.StoreKey, cdc codec.BinaryCodec) Migrator {
	return Migrator{key: key, cdc: cdc}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.key, m.cdc)
}
