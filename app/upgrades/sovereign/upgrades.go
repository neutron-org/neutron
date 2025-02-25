package v600

import (
	"context"
	"cosmossdk.io/math"
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	appparams "github.com/neutron-org/neutron/v5/app/params"
	revenuekeeper "github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v5/app/upgrades"
	harpoonkeeper "github.com/neutron-org/neutron/v5/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
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
			return vm, err
		}

		err = FuncAccounts(ctx, keepers.BankKeeper)
		if err != nil {
			return vm, err
		}

		err = SetupRewards(ctx, keepers.BankKeeper)
		if err != nil {
			return vm, err
		}

		// subscribe tracing contract for staking hooks to mirror staking power into the contract
		err = SetupTracking(ctx, keepers.HarpoonKeeper, keepers.WasmKeeper)
		if err != nil {
			return vm, err
		}

		// initial setup revenue module
		err = SetupRevenue(ctx, *keepers.RevenueKeeper)
		if err != nil {
			return vm, err
		}

		// move consensus from ICS validator (consumer module) to sovereign3 validators (staking module)
		err = DeICS(ctx, *keepers.StakingKeeper, *keepers.ConsumerKeeper, keepers.BankKeeper)
		if err != nil {
			return vm, err
		}

		// stake whole treasury through the Drop
		err = StakeWithDrop(ctx, *keepers.StakingKeeper, keepers.BankKeeper, keepers.WasmKeeper)
		if err != nil {
			return vm, err
		}

		err = SetupFeeMarket(ctx, keepers.FeeMarketKeeper)
		if err != nil {
			return vm, err
		}

		err = PinNewCodes(ctx, keepers.WasmKeeper)
		if err != nil {
			return vm, err
		}
		
		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

// TEST PURPOSES ONLY
func FuncAccounts(ctx context.Context, bk bankkeeper.Keeper) error {
	err := bk.MintCoins(ctx, "dex", sdk.NewCoins(sdk.Coin{
		Denom:  "untrn",
		Amount: math.NewInt(2_000_000_000_000),
	}))
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(MainDAOContractAddress)
	if err != nil {
		return err
	}
	err = bk.SendCoinsFromModuleToAccount(ctx, "dex", addr, sdk.NewCoins(sdk.Coin{
		Denom:  "untrn",
		Amount: math.NewInt(1_000_000_000_000),
	}))
	if err != nil {
		return err
	}

	//const (
	//	// neutron1jxxfkkxd9qfjzpvjyr9h3dy7l5693kx4y0zvay
	//	OperatorSk1 = "neutronvaloper1jxxfkkxd9qfjzpvjyr9h3dy7l5693kx47jm4mq"
	//	// neutron1tedsrwal9n2qlp6j3xcs3fjz9khx7z4reep8k3
	//	OperatorSk2 = "neutronvaloper1tedsrwal9n2qlp6j3xcs3fjz9khx7z4rryc7s4"
	//	// neutron1xdlvhs2l2wq0cc3eskyxphstns3348elwzvemh
	//	OperatorSk3 = "neutronvaloper1xdlvhs2l2wq0cc3eskyxphstns3348el5l4qan"
	//)
	vals := []string{
		"neutron1jxxfkkxd9qfjzpvjyr9h3dy7l5693kx4y0zvay",
		"neutron1tedsrwal9n2qlp6j3xcs3fjz9khx7z4reep8k3",
		"neutron1xdlvhs2l2wq0cc3eskyxphstns3348elwzvemh",
	}
	for _, a := range vals {
		addr, err := sdk.AccAddressFromBech32(a)
		if err != nil {
			return err
		}
		err = bk.SendCoinsFromModuleToAccount(ctx, "dex", addr, sdk.NewCoins(sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(1000000),
		}))
		if err != nil {
			return err
		}
	}
	return nil
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
			types.HOOK_TYPE_BEFORE_VALIDATOR_MODIFIED,
			types.HOOK_TYPE_AFTER_VALIDATOR_REMOVED,
			types.HOOK_TYPE_AFTER_VALIDATOR_BONDED,
			types.HOOK_TYPE_AFTER_VALIDATOR_BEGIN_UNBONDING,
			types.HOOK_TYPE_BEFORE_DELEGATION_CREATED,
			types.HOOK_TYPE_BEFORE_DELEGATION_SHARES_MODIFIED,
			types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED,
			types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED,
			types.HOOK_TYPE_BEFORE_VALIDATOR_SLASHED,
		}})

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
	daoBalance, err := bk.Balance(ctx, &types2.QueryBalanceRequest{
		Address: MainDAOContractAddress,
		Denom:   appparams.DefaultDenom,
	})
	if err != nil {
		return err
	}

	// TODO: The real amount will be defined later. Send half of DAO in test purposes
	rewardsAmount := daoBalance.Balance.Amount.QuoRaw(2)

	err = bk.SendCoins(
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

func SetupRevenue(ctx context.Context, rk revenuekeeper.Keeper) error {
	params := revenuetypes.Params{
		BaseCompensation:                  2500,
		BlocksPerformanceRequirement:      revenuetypes.DefaultBlocksPerformanceRequirement(),
		OracleVotesPerformanceRequirement: revenuetypes.DefaultOracleVotesPerformanceRequirement(),
		PaymentScheduleType: &revenuetypes.PaymentScheduleType{
			PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
				BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{
					BlocksPerPeriod: 600,
				},
			}},
		TwapWindow: 900,
	}
	srv := revenuekeeper.NewMsgServerImpl(&rk)
	_, err := srv.UpdateParams(ctx, &revenuetypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Params:    params,
	})
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

// PinNewCodes pins the new added codes
func PinNewCodes(ctx sdk.Context, wk *wasmkeeper.Keeper) error {
	wasmSrv := wasmkeeper.NewMsgServerImpl(wk)
	_, err := wasmSrv.PinCodes(ctx, &wasmTypes.MsgPinCodes{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		CodeIDs:   CodesToPin,
	})
	return err
}
