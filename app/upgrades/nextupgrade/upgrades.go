package nextupgrade

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/baseapp"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee/types"
	v6 "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	auctionkeeper "github.com/skip-mev/block-sdk/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/neutron-org/neutron/app/upgrades"
	contractmanagerkeeper "github.com/neutron-org/neutron/x/contractmanager/keeper"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
	crontypes "github.com/neutron-org/neutron/x/cron/types"
	feeburnerkeeper "github.com/neutron-org/neutron/x/feeburner/keeper"
	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"
	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	storeKeys upgrades.StoreKeys,
	codec codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Migrating channel capability...")
		// https://github.com/cosmos/ibc-go/blob/v7.0.1/docs/migrations/v5-to-v6.md#upgrade-proposal
		if err := v6.MigrateICS27ChannelCapability(ctx, codec, storeKeys.GetKey(capabilitytypes.StoreKey), keepers.CapabilityKeeper, interchaintxstypes.ModuleName); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migrating cron module parameters...")
		if err := migrateCronParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(crontypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating feerefunder module parameters...")
		if err := migrateFeeRefunderParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(feerefundertypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating tokenfactory module parameters...")
		if err := migrateTokenFactoryParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(tokenfactorytypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating feeburner module parameters...")
		if err := migrateFeeburnerParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(feeburnertypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating interchainqueries module parameters...")
		if err := migrateInterchainQueriesParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(icqtypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating interchaintxs module parameters...")
		if err := setInterchainTxsParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(interchaintxstypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting pob params...")
		err = setAuctionParams(ctx, keepers.FeeBurnerKeeper, keepers.AuctionKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting sudo callback limit...")
		err = setContractManagerParams(ctx, keepers.ContractManager)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating globalminfees module parameters...")
		err = migrateGlobalFees(ctx, keepers)
		if err != nil {
			ctx.Logger().Error("failed to migrate GlobalFees", "err", err)
			return vm, err
		}

		ctx.Logger().Info("Updating ccv reward denoms...")
		err = migrateRewardDenoms(ctx, keepers)
		if err != nil {
			ctx.Logger().Error("failed to update reward denoms", "err", err)
			return vm, err
		}

		ctx.Logger().Info("migrating adminmodule...")
		err = migrateAdminModule(ctx, keepers)
		if err != nil {
			ctx.Logger().Error("failed to migrate admin module", "err", err)
			return vm, err
		}

		ctx.Logger().Info("Migrating consensus params...")
		migrateConsensusParams(ctx, keepers.ParamsKeeper, keepers.ConsensusKeeper)

		ctx.Logger().Info("Upgrade complete")
		return vm, nil
	}
}

func setAuctionParams(ctx sdk.Context, feeBurnerKeeper *feeburnerkeeper.Keeper, auctionKeeper auctionkeeper.Keeper) error {
	treasury := feeBurnerKeeper.GetParams(ctx).TreasuryAddress
	_, data, err := bech32.DecodeAndConvert(treasury)
	if err != nil {
		return err
	}

	auctionParams := auctiontypes.Params{
		MaxBundleSize:          2,
		EscrowAccountAddress:   data,
		ReserveFee:             sdk.Coin{Denom: "untrn", Amount: sdk.NewInt(1_000_000)},
		MinBidIncrement:        sdk.Coin{Denom: "untrn", Amount: sdk.NewInt(1_000_000)},
		FrontRunningProtection: true,
		ProposerFee:            math.LegacyNewDecWithPrec(25, 2),
	}
	return auctionKeeper.SetParams(ctx, auctionParams)
}

func setContractManagerParams(ctx sdk.Context, keeper contractmanagerkeeper.Keeper) error {
	cmParams := contractmanagertypes.Params{
		SudoCallGasLimit: contractmanagertypes.DefaultSudoCallGasLimit,
	}
	return keeper.SetParams(ctx, cmParams)
}

func migrateCronParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams crontypes.Params
	subspace, _ := paramsKeepers.GetSubspace(crontypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(crontypes.ParamsKey, bz)
	return nil
}

func migrateFeeRefunderParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams feerefundertypes.Params
	subspace, _ := paramsKeepers.GetSubspace(feerefundertypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(feerefundertypes.ParamsKey, bz)
	return nil
}

func migrateTokenFactoryParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams tokenfactorytypes.Params
	subspace, _ := paramsKeepers.GetSubspace(tokenfactorytypes.StoreKey)
	subspace.Set(ctx, tokenfactorytypes.KeyDenomCreationGasConsume, uint64(0))
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(tokenfactorytypes.ParamsKey, bz)
	return nil
}

func migrateFeeburnerParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams feeburnertypes.Params
	subspace, _ := paramsKeepers.GetSubspace(feeburnertypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(feeburnertypes.ParamsKey, bz)
	return nil
}

func migrateInterchainQueriesParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams icqtypes.Params
	subspace, _ := paramsKeepers.GetSubspace(icqtypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(icqtypes.ParamsKey, bz)
	return nil
}

func setInterchainTxsParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams interchaintxstypes.Params
	subspace, _ := paramsKeepers.GetSubspace(interchaintxstypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)
	currParams.RegisterFee = interchaintxstypes.DefaultRegisterFee

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(interchaintxstypes.ParamsKey, bz)
	return nil
}

func migrateGlobalFees(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Implementing GlobalFee Params...")

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

	keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes, &[]string{})

	ctx.Logger().Info("Bypass min fee msg types was set successfully")

	keepers.GlobalFeeSubspace.Set(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &types.DefaultmaxTotalBypassMinFeeMsgGasUsage)

	ctx.Logger().Info("Max total bypass min fee msg gas usage set successfully")

	return nil
}

func migrateRewardDenoms(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Migrating reword denoms...")

	if !keepers.CcvConsumerSubspace.Has(ctx, ccvconsumertypes.KeyRewardDenoms) {
		return fmt.Errorf("key_reward_denoms param not found")
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

func migrateAdminModule(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Migrating admin module...")

	keepers.AdminModule.SetProposalID(ctx, 1)

	ctx.Logger().Info("Finished migrating admin module")

	return nil
}

func migrateConsensusParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, keeper *consensuskeeper.Keeper) {
	baseAppLegacySS := paramsKeepers.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	baseapp.MigrateParams(ctx, baseAppLegacySS, keeper)
}
