package nextupgrade

import (
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

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
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	builderkeeper "github.com/skip-mev/pob/x/builder/keeper"
	buildertypes "github.com/skip-mev/pob/x/builder/types"

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
		err = setPobParams(ctx, keepers.FeeBurnerKeeper, keepers.BuilderKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting sudo callback limit...")
		err = setContractManagerParams(ctx, keepers.ContractManagerKeeper)
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

		err = migrateAdminModule(ctx, keepers)
		if err != nil {
			ctx.Logger().Error("failed to migrate admin module", "err", err)
			return vm, err
		}

		if plan.Info != "testing_turn_off_contract_migrations" {
			err = migrateDaoContracts(ctx, keepers)
			if err != nil {
				ctx.Logger().Error("failed to migrate DAO contracts", "err", err)
				return vm, err
			}
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

	keepers.AdminModuleKeeper.SetProposalID(ctx, 1)

	ctx.Logger().Info("Finished migrating admin module")

	return nil
}

var (
	cwdCore                   uint64 = 246
	cwdProposalSingle         uint64 = 247
	cwdPreProposeSingle       uint64 = 248
	cwdProposalMultiple       uint64 = 249
	cwdPreProposeMultiple     uint64 = 250
	cwdPreProposeOverrule     uint64 = 251
	neutronVotingRegistry     uint64 = 252
	neutronVault              uint64 = 253
	creditsVault              uint64 = 254
	lockdropVault             uint64 = 255
	vestingLPVault            uint64 = 256
	investorsVestingVault     uint64 = 257
	cwdSubdaoCore             uint64 = 258
	cwdSubdaoPreProposeSingle uint64 = 259
	cwdSubdaoProposalSingle   uint64 = 260
	// var uint64 cwdSubdaoTimelockSingle = 261 # unused now: FIXME: use?
	neutronDistribution uint64 = 262
	neutronReserve      uint64 = 263
)

var (
	fromCompatibleMsg string = "{ \"from_compatible\": {} }"
	emptyMsg                 = "{}"
)

// var uint64 cwdSubdaoPreProposeSingleNoTimelock = 264 # unused now: FIXME: use?

func migrateDaoContracts(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Migrating dao contracts...")

	// main dao # cwd_core.wasm
	if err := migrateContract(ctx, keepers, "neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff", cwdCore, emptyMsg); err != nil {
		return err
	}
	// single proposal # cwd_proposal_single.wasm
	if err := migrateContract(ctx, keepers, "neutron1436kxs0w2es6xlqpp9rd35e3d0cjnw4sv8j3a7483sgks29jqwgshlt6zh", cwdProposalSingle, emptyMsg); err != nil {
		return err
	}
	// single pre proposal # cwd_pre_propose_single.wasm
	if err := migrateContract(ctx, keepers, "neutron1hulx7cgvpfcvg83wk5h96sedqgn72n026w6nl47uht554xhvj9nsgs8v0z", cwdPreProposeSingle, emptyMsg); err != nil {
		return err
	}
	// multiple proposal # cwd_proposal_multiple.wasm
	if err := migrateContract(ctx, keepers, "neutron1pvrwmjuusn9wh34j7y520g8gumuy9xtl3gvprlljfdpwju3x7ucsj3fj40", cwdProposalMultiple, fromCompatibleMsg); err != nil {
		return err
	}
	// multiple pre proposal # cwd_pre_propose_multiple.wasm
	if err := migrateContract(ctx, keepers, "neutron1up07dctjqud4fns75cnpejr4frmjtddzsmwgcktlyxd4zekhwecqt2h8u6", cwdPreProposeMultiple, emptyMsg); err != nil {
		return err
	}
	// overrule proposal # cwd_pre_propose_single.wasm
	if err := migrateContract(ctx, keepers, "neutron12pwnhtv7yat2s30xuf4gdk9qm85v4j3e6p44let47pdffpklcxlq56v0te", cwdPreProposeSingle, emptyMsg); err != nil {
		return err
	}
	// overrule pre proposal # cwd_pre_propose_overrule.wasm
	if err := migrateContract(ctx, keepers, "neutron1w798gp0zqv3s9hjl3jlnwxtwhykga6rn93p46q2crsdqhaj3y4gsum0096", cwdPreProposeOverrule, emptyMsg); err != nil {
		return err
	}
	// voting registry # neutron_voting_registry.wasm
	if err := migrateContract(ctx, keepers, "neutron1f6jlx7d9y408tlzue7r2qcf79plp549n30yzqjajjud8vm7m4vdspg933s", neutronVotingRegistry, emptyMsg); err != nil {
		return err
	}
	// ntrn vault # neutron_vault.wasm
	if err := migrateContract(ctx, keepers, "neutron1qeyjez6a9dwlghf9d6cy44fxmsajztw257586akk6xn6k88x0gus5djz4e", neutronVault, emptyMsg); err != nil {
		return err
	}
	// credits vault # credits_vault.wasm
	if err := migrateContract(ctx, keepers, "neutron1rxwzsw37ulveefk20575mlxl3hzhzv9k46c8gklfkt4g2vk4w3tse8usrs", creditsVault, emptyMsg); err != nil {
		return err
	}
	// lockdrop vault # lockdrop_vault.wasm
	if err := migrateContract(ctx, keepers, "neutron1f8gs4rp232ngyta3g2efwfkznymvv85du7qm9y0mhvjxpp3cq68qgquudm", lockdropVault, emptyMsg); err != nil {
		return err
	}
	// lp vesting vault # vesting_lp_vault.wasm
	if err := migrateContract(ctx, keepers, "neutron1adavpfxyp5kgs3zp0n0vkc37qakeh5eqwxqxzysgg0ahlx82rmsqp4rnz8", vestingLPVault, emptyMsg); err != nil {
		return err
	}
	// investors vesting vault # investors_vesting_vault.wasm
	if err := migrateContract(ctx, keepers, "neutron1dmd56h7hlevuwssp203fgc2uh0qdtwep2m735fzksuavgq3naslqp0ehvx", investorsVestingVault, emptyMsg); err != nil {
		return err
	}
	// security subdao # cwd_subdao_core.wasm
	if err := migrateContract(ctx, keepers, "neutron1fuyxwxlsgjkfjmxfthq8427dm2am3ya3cwcdr8gls29l7jadtazsuyzwcc", cwdSubdaoCore, fromCompatibleMsg); err != nil {
		return err
	}
	// security subdao single proposal # cwd_subdao_proposal_single.wasm
	if err := migrateContract(ctx, keepers, "neutron15m728qxvtat337jdu2f0uk6pu905kktrxclgy36c0wd822tpxcmqvnrurt", cwdSubdaoProposalSingle, emptyMsg); err != nil {
		return err
	}
	// security subdao single pre propose # cwd_subdao_pre_propose_single.wasm
	if err := migrateContract(ctx, keepers, "neutron1zjd5lwhch4ndnmayqxurja4x5y5mavy9ktrk6fzsyzan4wcgawnqjk5g26", cwdSubdaoPreProposeSingle, emptyMsg); err != nil {
		return err
	}
	// grants subdao cwd_subdao_core.wasm
	if err := migrateContract(ctx, keepers, "neutron1zjdv3u6svlazlydmje2qcp44yqkt0059chz8gmyl5yrklmgv6fzq9chelu", cwdSubdaoCore, fromCompatibleMsg); err != nil {
		return err
	}
	// grants subdao single proposal # cwd_subdao_proposal_single.wasm
	if err := migrateContract(ctx, keepers, "neutron14n7jt2qkngxtgr7dgdt50g4xn2a29llz79h9y25lrsqyxrwmngmsmt9kta", cwdSubdaoProposalSingle, emptyMsg); err != nil {
		return err
	}
	// grants subdao single pre propose # cwd_subdao_pre_propose_single.wasm
	if err := migrateContract(ctx, keepers, "neutron1s0fjev2pmgyaj0uthszzp3tpx59yp2p07vwhj0467sl9j343dk9qss6x9w", cwdSubdaoPreProposeSingle, emptyMsg); err != nil {
		return err
	}
	// distribution neutron_distribution.wasm
	if err := migrateContract(ctx, keepers, "neutron1dk9c86h7gmvuaq89cv72cjhq4c97r2wgl5gyfruv6shquwspalgq5u7sy5", neutronDistribution, emptyMsg); err != nil {
		return err
	}
	// reserve neutron_reserve.wasm
	if err := migrateContract(ctx, keepers, "neutron13we0myxwzlpx8l5ark8elw5gj5d59dl6cjkzmt80c5q5cv5rt54qvzkv2a", neutronReserve, emptyMsg); err != nil {
		return err
	}

	ctx.Logger().Info("Finished migrating dao contracts")

	return nil
}

func migrateContract(ctx sdk.Context, keepers *upgrades.UpgradeKeepers, contractAddress string, codeId uint64, msg string) error {
	_, err := keepers.WasmMsgServer.MigrateContract(sdk.WrapSDKContext(ctx), &wasmtypes.MsgMigrateContract{
		Sender:   "neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff", // main dao
		Contract: contractAddress,
		CodeID:   codeId,
		Msg:      []byte(msg),
	})
	return err
}
