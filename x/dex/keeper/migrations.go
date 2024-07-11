package keeper

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// // Migrate2to3 migrates from version 2 to 3.
// func (m Migrator) Migrate2to3(ctx sdk.Context) error {
//	return v3.MigrateStore(ctx, m.keeper.cdc, m.keeper.storeKey)
// }

// // Migrate2to3 migrates from version 3 to 4.
// func (m Migrator) Migrate3to4(ctx sdk.Context) error {
//	return v4.MigrateStore(ctx, m.keeper.cdc, m.keeper.storeKey)
// }
