package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v3 "github.com/neutron-org/neutron/v4/x/dex/migrations/v3"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.cdc, m.keeper.storeKey)
}
