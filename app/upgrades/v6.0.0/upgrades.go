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

		err = FundDNTRNLiqProvider(ctx, keepers.BankKeeper)
		if err != nil {
			return vm, fmt.Errorf("FundDNTRNLiqProvider failed: %w", err)
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

func SetupRevenue(ctx context.Context, rk revenuekeeper.Keeper, bk bankkeeper.Keeper) error {
	params := revenuetypes.Params{
		BaseCompensation:                  2500,
		BlocksPerformanceRequirement:      revenuetypes.DefaultBlocksPerformanceRequirement(),
		OracleVotesPerformanceRequirement: revenuetypes.DefaultOracleVotesPerformanceRequirement(),
		PaymentScheduleType: &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
				BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{
					BlocksPerPeriod: 600,
				},
			},
		},
		TwapWindow: 900,
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

func FundDNTRNLiqProvider(ctx context.Context, bk bankkeeper.Keeper) error {
	amount := math.NewInt(dntrnNtrnLiqamount)

	err := bk.SendCoins(
		ctx,
		sdk.MustAccAddressFromBech32(MainDAOContractAddress),
		sdk.MustAccAddressFromBech32(dntrnNtrnLiqprovider),
		sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, amount)),
	)
	if err != nil {
		return err
	}
	return nil
}
