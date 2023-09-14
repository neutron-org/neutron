package sdk47

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v6 "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
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
	builderkeeper "github.com/skip-mev/pob/x/builder/keeper"
	buildertypes "github.com/skip-mev/pob/x/builder/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	storeKeys upgrades.StoreKeys,
	codec codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migrating channel capability...")
		// https://github.com/cosmos/ibc-go/blob/v7.0.1/docs/migrations/v5-to-v6.md#upgrade-proposal
		if err := v6.MigrateICS27ChannelCapability(ctx, codec, storeKeys.GetKey(capabilitytypes.StoreKey), keepers.CapabilityKeeper, interchaintxstypes.ModuleName); err != nil {
			return nil, err
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
		if err := migrateInterchainTxsParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(interchaintxstypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting pob params...")
		err = setPobParams(ctx, keepers.FeeBurnerKeeper, keepers.BuilderKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting sudo callback limit...")
		err = setContractManagerParams(ctx, keepers.ContractManager)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, nil
	}
}

func setPobParams(ctx sdk.Context, feeBurnerKeeper *feeburnerkeeper.Keeper, builderKeeper builderkeeper.Keeper) error {
	treasury := feeBurnerKeeper.GetParams(ctx).TreasuryAddress
	_, data, err := bech32.DecodeAndConvert(treasury)
	if err != nil {
		return err
	}

	builderParams := buildertypes.Params{
		MaxBundleSize:          2,
		EscrowAccountAddress:   data,
		ReserveFee:             sdk.Coin{Denom: "untrn", Amount: sdk.NewInt(1_000_000)},
		MinBidIncrement:        sdk.Coin{Denom: "untrn", Amount: sdk.NewInt(1_000_000)},
		FrontRunningProtection: true,
		ProposerFee:            math.LegacyNewDecWithPrec(25, 2),
	}
	return builderKeeper.SetParams(ctx, builderParams)
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

func migrateInterchainTxsParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
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
