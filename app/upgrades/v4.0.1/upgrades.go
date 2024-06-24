package v400

import (
	"context"
	"fmt"
	"sort"

	appparams "github.com/neutron-org/neutron/v4/app/params"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	dynamicfeeskeeper "github.com/neutron-org/neutron/v4/x/dynamicfees/keeper"
	dynamicfeestypes "github.com/neutron-org/neutron/v4/x/dynamicfees/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	slinkyconstants "github.com/skip-mev/slinky/cmd/constants"

	"github.com/neutron-org/neutron/v4/app/upgrades"
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

		if ctx.ChainID() != "pion-1" {
			// skip the migration set since
			// the following changes already were applied on `pion-1` chain during `v4.0.1` release
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
var NtrnPrices = sdk.NewDecCoins(
	// Token,Denom,TWAP30D (USD),Price in NTRN (30d TWAP of DENOM / 30d TWAP of NTRN),Price + 30% premium
	// wstETH,factory/neutron1ug740qrkquxzrk2hh29qrlx3sktkfml3je7juusc2te7xmvsscns0n2wry/wstETH,3616.34,5779.670768739013,7513.571999360713
	sdk.NewDecCoinFromDec("factory/neutron1ug740qrkquxzrk2hh29qrlx3sktkfml3je7juusc2te7xmvsscns0n2wry/wstETH", math.LegacyMustNewDecFromStr("7513.571999360713")),

	// stATOM,ibc/B7864B03E1B9FD4F049243E92ABD691586F682137037A9F3FCA5222815620B3C,10.268,16.410420329231,21.333546427991
	sdk.NewDecCoinFromDec("ibc/B7864B03E1B9FD4F049243E92ABD691586F682137037A9F3FCA5222815620B3C", math.LegacyMustNewDecFromStr("21.333546427991")),

	// stTIA,ibc/6569E05DEE32B339D9286A52BE33DFCEFC97267F23EF9CFDE0C055140967A9A5,8.52,13.61674924085,17.70177401312
	sdk.NewDecCoinFromDec("ibc/6569E05DEE32B339D9286A52BE33DFCEFC97267F23EF9CFDE0C055140967A9A5", math.LegacyMustNewDecFromStr("17.70177401312")),

	// stkATOM,ibc/3649CE0C8A2C79048D8C6F31FF18FA69C9BC7EB193512E0BD03B733011290445,9.656,15.432315806297,20.062010548187
	sdk.NewDecCoinFromDec("ibc/3649CE0C8A2C79048D8C6F31FF18FA69C9BC7EB193512E0BD03B733011290445", math.LegacyMustNewDecFromStr("20.062010548187")),

	// USDC.noble,ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81,1,1.598210004795,2.077673006235
	sdk.NewDecCoinFromDec("ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81", math.LegacyMustNewDecFromStr("2.077673006235")),

	// USDC.axl,ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349,1,1.598210004795,2.077673006235
	sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", math.LegacyMustNewDecFromStr("2.077673006235")),

	// USDT,ibc/57503D7852EF4E1899FE6D71C5E81D7C839F76580F86F21E39348FC2BC9D7CE2,1,1.598210004795,2.077673006235
	sdk.NewDecCoinFromDec("ibc/57503D7852EF4E1899FE6D71C5E81D7C839F76580F86F21E39348FC2BC9D7CE2", math.LegacyMustNewDecFromStr("2.077673006235")),

	// http://astro.cw/,ibc/5751B8BCDA688FD0A8EC0B292EEF1CDEAB4B766B63EC632778B196D317C40C3A,0.0693,0.110755953332,0.143982739322
	sdk.NewDecCoinFromDec("ibc/5751B8BCDA688FD0A8EC0B292EEF1CDEAB4B766B63EC632778B196D317C40C3A", math.LegacyMustNewDecFromStr("0.143982739322")),

	// ASTRO,factory/neutron1ffus553eet978k024lmssw0czsxwr97mggyv85lpcsdkft8v9ufsz3sa07/astro,0.0693,0.110755953332,0.143982739322
	sdk.NewDecCoinFromDec("factory/neutron1ffus553eet978k024lmssw0czsxwr97mggyv85lpcsdkft8v9ufsz3sa07/astro", math.LegacyMustNewDecFromStr("0.143982739322")),

	// MARS,ibc/9598CDEB7C6DB7FC21E746C8E0250B30CD5154F39CA111A9D4948A4362F638BD,0.0652,0.104203292313,0.135464280003
	sdk.NewDecCoinFromDec("ibc/9598CDEB7C6DB7FC21E746C8E0250B30CD5154F39CA111A9D4948A4362F638BD", math.LegacyMustNewDecFromStr("0.135464280003")),

	// APOLLO,factory/neutron154gg0wtm2v4h9ur8xg32ep64e8ef0g5twlsgvfeajqwghdryvyqsqhgk8e/APOLLO,0.06079,0.097155186191,0.126301742051
	sdk.NewDecCoinFromDec("factory/neutron154gg0wtm2v4h9ur8xg32ep64e8ef0g5twlsgvfeajqwghdryvyqsqhgk8e/APOLLO", math.LegacyMustNewDecFromStr("0.126301742051")),

	// TIA,ibc/773B4D0A3CD667B2275D5A4A7A2F0909C0BA0F4059C0B9181E680DDF4965DCC7,8.52,13.61674924085,17.70177401312
	sdk.NewDecCoinFromDec("ibc/773B4D0A3CD667B2275D5A4A7A2F0909C0BA0F4059C0B9181E680DDF4965DCC7", math.LegacyMustNewDecFromStr("17.70177401312")),

	// ATOM,ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9,7.559,12.080869426243,15.705130254103
	sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", math.LegacyMustNewDecFromStr("15.705130254103")),

	// axlWETH,ibc/A585C2D15DCD3B010849B453A2CFCB5E213208A5AB665691792684C26274304D,3616.34,5779.670768739013,7513.571999360713
	sdk.NewDecCoinFromDec("ibc/A585C2D15DCD3B010849B453A2CFCB5E213208A5AB665691792684C26274304D", math.LegacyMustNewDecFromStr("7513.571999360713")),

	// OSMO,ibc/376222D6D9DAE23092E29740E56B758580935A6D77C24C2ABD57A6A78A1F3955,0.6699,1.070640882212,1.391833146872
	sdk.NewDecCoinFromDec("ibc/376222D6D9DAE23092E29740E56B758580935A6D77C24C2ABD57A6A78A1F3955", math.LegacyMustNewDecFromStr("1.391833146872")),

	// DYM,ibc/4A6A46D4263F2ED3DCE9CF866FE15E6903FB5E12D87EB8BDC1B6B1A1E2D397B4,2.329,3.722231101167,4.838900431527
	sdk.NewDecCoinFromDec("ibc/4A6A46D4263F2ED3DCE9CF866FE15E6903FB5E12D87EB8BDC1B6B1A1E2D397B4", math.LegacyMustNewDecFromStr("4.838900431527")),

	// DYDX,ibc/2CB87BCE0937B1D1DFCEE79BE4501AAF3C265E923509AEAC410AD85D27F35130,1.71,2.732939108199,3.552820840659
	sdk.NewDecCoinFromDec("ibc/2CB87BCE0937B1D1DFCEE79BE4501AAF3C265E923509AEAC410AD85D27F35130", math.LegacyMustNewDecFromStr("3.552820840659")),

	// KUJI,ibc/1053E271314D36FECBC915B51474F8B3962597CE88FF3E4A74795B0E3F367A8B,1.22,1.949816205849,2.534761067589
	sdk.NewDecCoinFromDec("ibc/1053E271314D36FECBC915B51474F8B3962597CE88FF3E4A74795B0E3F367A8B", math.LegacyMustNewDecFromStr("2.534761067589")),

	// STARS,ibc/A139C0E0B5E87CBA8EAEEB12B9BEE13AC7C814CFBBFA87BBCADD67E31003466C,0.0138,0.022055298066,0.028671887496
	sdk.NewDecCoinFromDec("ibc/A139C0E0B5E87CBA8EAEEB12B9BEE13AC7C814CFBBFA87BBCADD67E31003466C", math.LegacyMustNewDecFromStr("0.028671887496")),

	// AXL,ibc/C0E66D1C81D8AAF0E6896E05190FDFBC222367148F86AC3EA679C28327A763CD,0.802,1.281764423845,1.666293750985
	sdk.NewDecCoinFromDec("ibc/C0E66D1C81D8AAF0E6896E05190FDFBC222367148F86AC3EA679C28327A763CD", math.LegacyMustNewDecFromStr("1.666293750985")),

	// STRD,ibc/3552CECB7BCE1891DB6070D37EC6E954C972B1400141308FCD85FD148BD06DE5,1.542,2.464439827393,3.203771775613
	sdk.NewDecCoinFromDec("ibc/3552CECB7BCE1891DB6070D37EC6E954C972B1400141308FCD85FD148BD06DE5", math.LegacyMustNewDecFromStr("3.203771775613")),
)

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

func setFeeMarketParams(ctx sdk.Context, feemarketKeeper *feemarketkeeper.Keeper) error {
	feemarketParams := feemarkettypes.Params{
		Alpha:               math.LegacyZeroDec(),
		Beta:                math.LegacyZeroDec(),
		Delta:               math.LegacyZeroDec(),
		MinBaseGasPrice:     math.LegacyMustNewDecFromStr("0.0053"),
		MinLearningRate:     math.LegacyMustNewDecFromStr("0.08"),
		MaxLearningRate:     math.LegacyMustNewDecFromStr("0.08"),
		MaxBlockUtilization: 30_000_000,
		Window:              1,
		FeeDenom:            appparams.DefaultDenom,
		Enabled:             true,
		DistributeFees:      true,
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

	oldParams.Params.Version = &comettypes.VersionParams{App: 0}
	if err := consensusKeeper.ParamsStore.Set(ctx, *oldParams.Params); err != nil {
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
