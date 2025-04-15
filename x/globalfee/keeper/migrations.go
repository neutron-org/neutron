package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v3 "github.com/neutron-org/neutron/v6/x/globalfee/migrations/v3"

	v2 "github.com/neutron-org/neutron/v6/x/globalfee/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	globalfeeSubspace paramtypes.Subspace
	cdc               codec.BinaryCodec
	storeKey          storetypes.StoreKey
}

// NewMigrator returns a new Migrator.
func NewMigrator(cdc codec.BinaryCodec, globalfeeSubspace paramtypes.Subspace, storeKey storetypes.StoreKey) Migrator {
	return Migrator{globalfeeSubspace: globalfeeSubspace, storeKey: storeKey, cdc: cdc}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.globalfeeSubspace)
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.cdc, m.globalfeeSubspace, m.storeKey)
}
