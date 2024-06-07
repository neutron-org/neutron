package v400

import (
	"context"
	"fmt"
	"sort"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	"github.com/neutron-org/neutron/v4/app/params"
	dynamicfeeskeeper "github.com/neutron-org/neutron/v4/x/dynamicfees/keeper"
	dynamicfeestypes "github.com/neutron-org/neutron/v4/x/dynamicfees/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	"github.com/neutron-org/neutron/v4/app/upgrades"
	slinkyconstants "github.com/skip-mev/slinky/cmd/constants"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Setting consensus params...")
		err = enableVoteExtensions(ctx, keepers.ConsensusKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting marketmap params...")
		err = setMarketMapParams(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting dynamicfees/feemarket params...")
		err = setFeeMarketParams(ctx, keepers.FeeMarketKeeper)
		if err != nil {
			return nil, err
		}

		err = setDynamicFeesParams(ctx, keepers.DynamicfeesKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting marketmap and oracle state...")
		err = setMarketState(ctx, keepers.MarketmapKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func setMarketMapParams(ctx sdk.Context, marketmapKeeper *marketmapkeeper.Keeper) error {
	marketmapParams := marketmaptypes.Params{
		MarketAuthorities: []string{authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(), MarketMapAuthorityMultisig},
		Admin:             authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	}
	return marketmapKeeper.SetParams(ctx, marketmapParams)
}

// NtrnPrices describes prices of any token in NTRN for dynamic fee resolver
// TODO: determine actual prices
var NtrnPrices = sdk.NewDecCoins(sdk.NewDecCoin(params.DefaultDenom, math.OneInt().Mul(math.NewInt(100))))

func setDynamicFeesParams(ctx sdk.Context, dfKeeper *dynamicfeeskeeper.Keeper) error {
	dfParams := dynamicfeestypes.Params{
		NtrnPrices: NtrnPrices,
	}
	err := dfKeeper.SetParams(ctx, dfParams)
	if err != nil {
		return errors.Wrap(err, "failed to set dynamic fees params")
	}

	return nil
}

// TODO: add a test for the migrations: check that feemarket state is consistent with feemarket params
func setFeeMarketParams(ctx sdk.Context, feemarketKeeper *feemarketkeeper.Keeper) error {
	// TODO: set params values
	feemarketParams := feemarkettypes.Params{
		Alpha:               math.LegacyDec{},
		Beta:                math.LegacyDec{},
		Delta:               math.LegacyDec{},
		MinBaseGasPrice:     math.LegacyDec{},
		MinLearningRate:     math.LegacyDec{},
		MaxLearningRate:     math.LegacyDec{},
		MaxBlockUtilization: 0,
		Window:              0,
		FeeDenom:            "",
		Enabled:             false,
		DistributeFees:      false,
	}
	feemarketState := feemarkettypes.NewState(feemarketParams.Window, feemarketParams.MinBaseGasPrice, feemarketParams.MinLearningRate)
	err := feemarketKeeper.SetParams(ctx, feemarketParams)
	if err != nil {
		return errors.Wrap(err, "failed to set feemarket params")
	}
	err = feemarketKeeper.SetState(ctx, feemarketState)
	if err != nil {
		return errors.Wrap(err, "failed to set feemarket state")
	}

	return nil
}

func setMarketState(ctx sdk.Context, mmKeeper *marketmapkeeper.Keeper) error {
	markets := marketMapToDeterministicallyOrderedMarkets(slinkyconstants.CoreMarketMap)
	for _, market := range markets {
		if err := mmKeeper.CreateMarket(ctx, market); err != nil {
			return err
		}

		if err := mmKeeper.Hooks().AfterMarketCreated(ctx, market); err != nil {
			return err
		}

	}
	return nil
}

func marketMapToDeterministicallyOrderedMarkets(mm marketmaptypes.MarketMap) []marketmaptypes.Market {
	markets := make([]marketmaptypes.Market, 0, len(mm.Markets))
	for _, market := range mm.Markets {
		markets = append(markets, market)
	}

	// order the markets alphabetically by their ticker.String()
	sort.Slice(markets, func(i, j int) bool {
		return markets[i].Ticker.String() < markets[j].Ticker.String()
	})

	return markets
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
