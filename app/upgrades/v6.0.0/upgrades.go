package v600

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"

	appparams "github.com/neutron-org/neutron/v6/app/params"
	dynamicfeeskeeper "github.com/neutron-org/neutron/v6/x/dynamicfees/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v6/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v6/app/upgrades"
	harpoonkeeper "github.com/neutron-org/neutron/v6/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

var Valopers = []string{
	"neutronvaloper1pcca7p0n4ghzgyyy8ccg4zax35wzqhdje3f2fx",
	"neutronvaloper1rc357mf6kmdh9ecdngnkckhw4usgrtnwqnwz0e",
	"neutronvaloper15xm63t0tjzd5synhf7jlxajkzvk6qsk9tjp0sa",
	"neutronvaloper15fv4jxqgfctgsgcd4w3j5plzpd7glepr3y84e5",
	"neutronvaloper1vual9khy5djv9aktha9kxxlthqcvzlx92agpv4",
	"neutronvaloper123pgsme75kwwxy0r8qr40kshr9vgy4048gma9c",
	"neutronvaloper1sqmy5d7gexn570v90rjrz0kt4d4z6wte0x055t",
	"neutronvaloper1q2ppatmd9yt060qxpn598srel2pg6a5036s48a",
	"neutronvaloper1vk6kq00r6tdg0eyvlgwgn3v5z59q2ndxpcmpq2",
}

/*
Test setup instructions - https://www.notion.so/hadron/deICS-testnet-setup-19885d6b9b10802fa08bd2b5effa06ae#19885d6b9b1080d79f3beefba2231c75
*/

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
			return vm, fmt.Errorf("RunMigrations failed: %w", err)
		}

		SetupDenomMetadata(ctx, keepers.BankKeeper)

		err = SetupRewards(ctx, keepers.BankKeeper)
		if err != nil {
			return vm, fmt.Errorf("SetupRewards failed: %w", err)
		}

		// subscribe tracing contract for staking hooks to mirror staking power into the contract
		err = SetupTracking(ctx, keepers.HarpoonKeeper, keepers.WasmKeeper)
		if err != nil {
			return vm, fmt.Errorf("SetupTracking failed: %w", err)
		}

		// initial setup revenue module
		err = SetupRevenue(ctx, *keepers.RevenueKeeper, keepers.BankKeeper)
		if err != nil {
			return vm, fmt.Errorf("SetupRevenue failed: %w", err)
		}

		// move consensus from ICS validator (consumer module) to sovereign3 validators (staking module)
		err = DeICS(ctx, *keepers.StakingKeeper, *keepers.ConsumerKeeper, keepers.BankKeeper)
		if err != nil {
			return vm, fmt.Errorf("DeICS failed: %w", err)
		}

		// stake whole treasury through the Drop
		err = StakeWithDrop(ctx, *keepers.StakingKeeper, keepers.BankKeeper, keepers.WasmKeeper)
		if err != nil {
			return vm, fmt.Errorf("StakeWithDrop failed: %w", err)
		}

		err = SetupFeeMarket(ctx, keepers.FeeMarketKeeper)
		if err != nil {
			return vm, fmt.Errorf("SetupFeeMarket failed: %w", err)
		}

		err = SetupDynamicfees(ctx, keepers.DynamicfeesKeeper)
		if err != nil {
			return vm, fmt.Errorf("SetupDynamicfees failed: %w", err)
		}

		err = FundValence(ctx, keepers.BankKeeper)
		if err != nil {
			return vm, fmt.Errorf("FundValence failed: %w", err)
		}

		err = FundLiqUSDCLPProvider(ctx, keepers.BankKeeper)
		if err != nil {
			return vm, fmt.Errorf("FundLiqUSDCLPProvider failed: %w", err)
		}

		err = PinNewCodes(ctx, keepers.WasmKeeper)
		if err != nil {
			return vm, fmt.Errorf("PinNewCodes failed: %w", err)
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

type VotingRegistryExecuteMsg struct {
	AddVotingVault AddVotingVaultMsg `json:"add_voting_vault"`
}

type AddVotingVaultMsg struct {
	NewVotingVaultContract string `json:"new_voting_vault_contract"`
}

func SetupTracking(ctx sdk.Context, harpoonKeeper *harpoonkeeper.Keeper, wasmKeeper *wasmkeeper.Keeper) error {
	harpoonKeeper.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: StakingTrackerContractAddress,
		Hooks: []types.HookType{
			types.HOOK_TYPE_AFTER_VALIDATOR_CREATED,
			types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
			types.HOOK_TYPE_AFTER_VALIDATOR_BEGIN_UNBONDING,
			types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED,
			types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED,
			types.HOOK_TYPE_BEFORE_VALIDATOR_SLASHED,
		},
	})

	addVaultMsg := VotingRegistryExecuteMsg{AddVotingVault: AddVotingVaultMsg{NewVotingVaultContract: StakingVaultContractAddress}}
	addVaultMsgBz, err := json.Marshal(addVaultMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal AddVotingVault msg to json: %w", err)
	}

	wasmSrv := wasmkeeper.NewMsgServerImpl(wasmKeeper)
	_, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: VotingRegistryContractAddress,
		Msg:      addVaultMsgBz,
		Funds:    nil,
	})
	if err != nil {
		return fmt.Errorf("failed to add Staking Vault in the Voting Registry: %w", err)
	}

	return nil
}

func SetupRewards(ctx context.Context, bk bankkeeper.Keeper) error {
	rewardsAmount := math.NewInt(RewardContract)

	err := bk.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(MainDAOContractAddress),
		sdk.MustAccAddressFromBech32(StakingRewardsContractAddress),
		sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, rewardsAmount)),
	)
	if err != nil {
		return err
	}

	return nil
}

func SetupDenomMetadata(ctx context.Context, bk bankkeeper.Keeper) {
	bk.SetDenomMetaData(ctx, banktypes.Metadata{
		Description: "The native staking token of the Neutron network",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appparams.DefaultDenom,
				Exponent: 0,
				Aliases:  []string{"microntrn"},
			},
			{
				Denom:    "ntrn",
				Exponent: appparams.DefaultDenomDecimals,
				Aliases:  []string{"NTRN"},
			},
		},
		Base:    appparams.DefaultDenom,
		Display: "ntrn",
		Name:    "Neutron",
		Symbol:  "NTRN",
	})
}

func SetupRevenue(ctx context.Context, rk revenuekeeper.Keeper, bk bankkeeper.Keeper) error {
	params := revenuetypes.Params{
		RewardAsset: revenuetypes.DefaultRewardAsset,
		RewardQuote: &revenuetypes.RewardQuote{
			Asset:  revenuetypes.DefaultRewardQuoteAsset,
			Amount: revenuetypes.DefaultRewardQuoteAmount,
		},
		BlocksPerformanceRequirement:      revenuetypes.DefaultBlocksPerformanceRequirement(),
		OracleVotesPerformanceRequirement: revenuetypes.DefaultOracleVotesPerformanceRequirement(),
		PaymentScheduleType: &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
				BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{
					BlocksPerPeriod: 200,
				},
			},
		},
		TwapWindow: 200,
	}
	srv := revenuekeeper.NewMsgServerImpl(&rk)
	_, err := srv.UpdateParams(ctx, &revenuetypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Params:    params,
	})
	if err != nil {
		return err
	}

	revenueAmount := math.NewInt(RevenueModule)

	err = bk.SendCoinsFromAccountToModule(
		ctx,
		sdk.MustAccAddressFromBech32(MainDAOContractAddress),
		revenuetypes.RevenueTreasuryPoolName,
		sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, revenueAmount)),
	)
	return err
}

func SetupFeeMarket(ctx context.Context, fk *feemarketkeeper.Keeper) error {
	params, err := fk.GetParams(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return err
	}

	params.SendTipToProposer = false
	err = fk.SetParams(sdk.UnwrapSDKContext(ctx), params)
	if err != nil {
		return err
	}

	return nil
}

func SetupDynamicfees(ctx sdk.Context, dfk *dynamicfeeskeeper.Keeper) error {
	params := dfk.GetParams(sdk.UnwrapSDKContext(ctx))
	params.NtrnPrices = append(params.NtrnPrices, sdk.DecCoin{
		Denom:  DropNtrnDenom,
		Amount: math.LegacyOneDec(),
	})
	if err := dfk.SetParams(sdk.UnwrapSDKContext(ctx), params); err != nil {
		return fmt.Errorf("failed to set dynamicfees params with NtrnPrices updated: %w", err)
	}

	return nil
}

// PinNewCodes pins the new added codes
func PinNewCodes(ctx sdk.Context, wk *wasmkeeper.Keeper) error {
	wasmSrv := wasmkeeper.NewMsgServerImpl(wk)
	_, err := wasmSrv.PinCodes(ctx, &wasmTypes.MsgPinCodes{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		CodeIDs:   CodesToPin,
	})
	return err
}

func FundValence(ctx context.Context, bk bankkeeper.Keeper) error {
	amount := math.NewInt(StakeWithValenceAmount)

	err := bk.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(MainDAOContractAddress),
		sdk.MustAccAddressFromBech32(ValenceStaker),
		sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, amount)),
	)
	if err != nil {
		return err
	}
	return nil
}

func FundLiqUSDCLPProvider(ctx context.Context, bk bankkeeper.Keeper) error {
	// query amount and transfer 100%
	daoBalanceBefore, err := bk.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: MainDAOContractAddress,
		Denom:   UsdcLpDenom,
	})
	if err != nil {
		return err
	}

	err = bk.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(MainDAOContractAddress),
		sdk.MustAccAddressFromBech32(UsdcLpReceiver),
		sdk.NewCoins(*daoBalanceBefore.Balance),
	)
	if err != nil {
		return err
	}
	return nil
}
