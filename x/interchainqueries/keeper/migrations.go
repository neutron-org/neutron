package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v3 "github.com/neutron-org/neutron/v6/x/interchainqueries/migrations/v3"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewMigrator returns a new Migrator.
func NewMigrator(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Migrator {
	return Migrator{storeKey: storeKey, cdc: cdc}
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateParams(ctx, m.cdc, m.storeKey)
}
