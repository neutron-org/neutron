package nextupgrade

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/x/globalfee/types"
	"github.com/neutron-org/neutron/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		fmt.Println("CreateUpgradeHandler 1")
		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migrating GlobalFee Params...")

		if keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyMinGasPrices) == false {
			return vm, errors.New("minimum_gas_prices param not found")
		}
		// global fee is empty set, set global fee to
		// 0.01untrn,0.005ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9,0.05ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349
		requiredGlobalFees := sdk.DecCoins{
			sdk.NewDecCoinFromDec("untrn", sdk.MustNewDecFromStr("0.01")),
			sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.MustNewDecFromStr("0.005")),
			sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", sdk.MustNewDecFromStr("0.05")),
		}
		ctx.Logger().Info("Empty min gas prices, set global fees")

		requiredGlobalFees = requiredGlobalFees.Sort()

		keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyMinGasPrices, &requiredGlobalFees)

		ctx.Logger().Info("global fees was set successfully")

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
