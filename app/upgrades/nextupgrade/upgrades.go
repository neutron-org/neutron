package nextupgrade

import (
	"errors"

	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/neutron-org/neutron/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	storeKeys upgrades.StoreKeys,
	cdc codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		err = migrateGlobalFees(ctx, keepers)
		if err != nil {
			ctx.Logger().Error("failed to migrate GlobalFees", "err", err)
			return vm, err
		}

		err = migrateRewardDenoms(ctx, keepers)
		if err != nil {
			ctx.Logger().Error("failed to migrate reward denoms", "err", err)
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}

func migrateGlobalFees(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Implementing GlobalFee Params...")

	if !keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		return errors.New("minimum_gas_prices param not found")
	}

	if !keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes) {
		return errors.New("bypass_min_fee_msg_types param not found")
	}

	if !keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage) {
		return errors.New("max_total_bypass_min_fee_msg_gas_usage param not found")
	}

	// global fee is empty set, set global fee to equal to 0.05 USD (for 200k of gas) in appropriate coin
	// As of June 22nd, 2023 this is
	// 0.9untrn,0.026ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9,0.25ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349
	requiredGlobalFees := sdk.DecCoins{
		sdk.NewDecCoinFromDec("untrn", sdk.MustNewDecFromStr("0.9")),
		sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.MustNewDecFromStr("0.026")),
		sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", sdk.MustNewDecFromStr("0.25")),
	}
	requiredGlobalFees = requiredGlobalFees.Sort()

	keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyMinGasPrices, &requiredGlobalFees)

	ctx.Logger().Info("Global fees was set successfully")

	defaultBypassFeeMessages := []string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
	}
	keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes, &defaultBypassFeeMessages)

	ctx.Logger().Info("Bypass min fee msg types was set successfully")

	keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &types.DefaultmaxTotalBypassMinFeeMsgGasUsage)

	ctx.Logger().Info("Max total bypass min fee msg gas usage set successfully")

	return nil
}

func migrateRewardDenoms(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Migrating reword denoms...")

	if !keepers.CcvConsumerSubspace.Has(ctx, ccvconsumertypes.KeyRewardDenoms) {
		return errors.New("key_reward_denoms param not found")
	}

	var denoms []string
	keepers.CcvConsumerSubspace.Get(ctx, ccvconsumertypes.KeyRewardDenoms, &denoms)

	// add new axlr usdc denom
	axlrDenom := "ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349"
	denoms = append(denoms, axlrDenom)

	keepers.CcvConsumerSubspace.Set(ctx, ccvconsumertypes.KeyRewardDenoms, &denoms)

	ctx.Logger().Info("Finished migrating reward denoms")

	return nil
}
