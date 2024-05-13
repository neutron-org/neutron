package v400

import (
	"context"
	"fmt"

	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"

	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v4/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Setting marketmap params...")
		err = setMarketMapParams(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting consensus params...")
		err = enableVoteExtensions(ctx, keepers.ConsensusKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func setMarketMapParams(ctx sdk.Context, marketmapKeeper *marketmapkeeper.Keeper) error {
	params, err := marketmapKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	marketmapParams := marketmaptypes.Params{
		MarketAuthority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Version:         params.Version,
	}
	return marketmapKeeper.SetParams(ctx, marketmapParams)
}

func enableVoteExtensions(ctx sdk.Context, consensusKeeper *consensuskeeper.Keeper) error {
	oldParams, err := consensusKeeper.Params(ctx, &types.QueryParamsRequest{})
	if err != nil {
		return err
	}

	// we need to enable VoteExtensions for Slinky
	oldParams.Params.Abci = &comettypes.ABCIParams{VoteExtensionsEnableHeight: ctx.BlockHeight() + 4}

	updateParamsMsg := types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Block:     oldParams.Params.Block,
		Evidence:  oldParams.Params.Evidence,
		Validator: oldParams.Params.Validator,
		Abci:      oldParams.Params.Abci,
	}

	_, err = consensusKeeper.UpdateParams(ctx, &updateParamsMsg)
	return err
}
