package v200

import (
	"cosmossdk.io/math"

	"fmt"
	feeburnerkeeper "github.com/neutron-org/neutron/v2/x/feeburner/keeper"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/neutron-org/neutron/v2/app/params"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee/types"
	v6 "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	auctionkeeper "github.com/skip-mev/block-sdk/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/neutron-org/neutron/v2/app/upgrades"
	contractmanagerkeeper "github.com/neutron-org/neutron/v2/x/contractmanager/keeper"
	contractmanagertypes "github.com/neutron-org/neutron/v2/x/contractmanager/types"
	crontypes "github.com/neutron-org/neutron/v2/x/cron/types"
	feeburnertypes "github.com/neutron-org/neutron/v2/x/feeburner/types"
	feerefundertypes "github.com/neutron-org/neutron/v2/x/feerefunder/types"
	icqtypes "github.com/neutron-org/neutron/v2/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v2/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v2/x/tokenfactory/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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
		if err := setInterchainTxsParams(ctx, keepers.ParamsKeeper, storeKeys.GetKey(interchaintxstypes.StoreKey), storeKeys.GetKey(wasmtypes.StoreKey), codec); err != nil {
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
	subspace.Set(ctx, tokenfactorytypes.KeyDenomCreationFee, sdk.NewCoins())
	subspace.Set(ctx, tokenfactorytypes.KeyFeeCollectorAddress, "")
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

	currParams.QueryDeposit = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(1_000_000)))

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(icqtypes.ParamsKey, bz)
	return nil
}

func setInterchainTxsParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey, wasmStoreKey storetypes.StoreKey, codec codec.Codec) error {
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

	wasmStore := ctx.KVStore(wasmStoreKey)
	bzWasm := wasmStore.Get(wasmtypes.KeySequenceCodeID)
	if bzWasm == nil {
		return fmt.Errorf("KeySequenceCodeID not found during the upgrade")
	}
	store.Set(interchaintxstypes.ICARegistrationFeeFirstCodeID, bzWasm)
	return nil
}

func migrateGlobalFees(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error { //nolint:unparam
	ctx.Logger().Info("Implementing GlobalFee Params...")

	// The average gas cost for an average transaction on Neutron should not go beyond 5 cents.
	// Users have three designated coins that can be used for gas: NTRN, ATOM, and axlUSDC
	// Assuming average transaction gas on Neutron consumer is ~250000 approximately, ATOM 30D TWAP is $8.4 and NTRN 30D TWAP is $0.36
	// we set minimum-gas-prices as per this formula:
	// ((0.05 * 10^(6)) / TOKEN_30d_TWAP) / AVG_GAS_PRICE
	requiredGlobalFees := sdk.DecCoins{
		sdk.NewDecCoinFromDec(params.DefaultDenom, sdk.MustNewDecFromStr("0.56")),
		sdk.NewDecCoinFromDec(AtomDenom, sdk.MustNewDecFromStr("0.02")),
		sdk.NewDecCoinFromDec(AxelarUsdcDenom, sdk.MustNewDecFromStr("0.2")),
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

func migrateAdminModule(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error { //nolint:unparam
	ctx.Logger().Info("Migrating admin module...")

	keepers.AdminModule.SetProposalID(ctx, 1)

	ctx.Logger().Info("Finished migrating admin module")

	return nil
}

func migrateConsensusParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, keeper *consensuskeeper.Keeper) {
	baseAppLegacySS := paramsKeepers.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	baseapp.MigrateParams(ctx, baseAppLegacySS, keeper)
}
