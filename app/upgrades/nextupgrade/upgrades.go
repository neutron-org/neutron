package nextupgrade

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/neutron-org/neutron/app/upgrades"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
)

func MigrateFailures(ctx sdk.Context, storeKeys upgrades.StoreKeys, cdc codec.Codec) error {
	ctx.Logger().Info("Migrating failures...")

	// fetch list of all old failures
	oldFailuresList := make([]contractmanagertypes.OldFailure, 0)
	iteratorStore := prefix.NewStore(ctx.KVStore(storeKeys.GetKey(contractmanagertypes.StoreKey)), contractmanagertypes.ContractFailuresKey)
	iterator := sdk.KVStorePrefixIterator(iteratorStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		var val contractmanagertypes.OldFailure
		cdc.MustUnmarshal(iterator.Value(), &val)
		oldFailuresList = append(oldFailuresList, val)
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// migrate
	store := ctx.KVStore(storeKeys.GetKey(contractmanagertypes.StoreKey))
	for _, oldItem := range oldFailuresList {
		failure := contractmanagertypes.Failure{
			Address: oldItem.Address,
			Id:      oldItem.Id,
			AckType: oldItem.AckType,
			Packet:  nil,
			Ack:     nil,
		}
		bz := cdc.MustMarshal(&failure)
		store.Set(contractmanagertypes.GetFailureKey(failure.Address, failure.Id), bz)
	}

	ctx.Logger().Info("Finished migrating failures")

	return nil
}

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

		ctx.Logger().Info("Implementing GlobalFee Params...")

		if !keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyMinGasPrices) {
			return vm, errors.New("minimum_gas_prices param not found")
		}

		if !keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes) {
			return vm, errors.New("bypass_min_fee_msg_types param not found")
		}

		if !keepers.GlobalFeeSubspace.Has(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage) {
			return vm, errors.New("max_total_bypass_min_fee_msg_gas_usage param not found")
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

		keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, types.DefaultmaxTotalBypassMinFeeMsgGasUsage)

		ctx.Logger().Info("Max total bypass min fee msg gas usage set successfully")

		err = MigrateFailures(ctx, storeKeys, cdc)
		if err != nil {
			ctx.Logger().Error("failed to migrate failures", "err", err)
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
