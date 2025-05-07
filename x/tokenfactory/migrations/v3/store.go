package v3

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type TokenFactoryKeeper interface {
	GetAllDenomsIterator(ctx sdk.Context) storetypes.Iterator
}

type BankKeeper interface {
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}

// MigrateStore performs in-place store migrations.
// The migration sets name, symbol and display for created denoms metadata
func MigrateStore(ctx sdk.Context, keeper TokenFactoryKeeper, bankKeeper BankKeeper) error {
	if err := migrateParams(ctx, keeper, bankKeeper); err != nil {
		return err
	}
	return nil
}

func migrateParams(ctx sdk.Context, keeper TokenFactoryKeeper, bankKeeper BankKeeper) error {
	ctx.Logger().Info("Migrating denoms metadata...")

	iterator := keeper.GetAllDenomsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Value())

		metadata, _ := bankKeeper.GetDenomMetaData(ctx, denom)

		metadata.Name = denom
		metadata.Symbol = denom
		metadata.Display = denom

		bankKeeper.SetDenomMetaData(ctx, metadata)
	}

	ctx.Logger().Info("Finished migrating denoms metadata")

	return nil
}
