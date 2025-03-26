package v600

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	appparams "github.com/neutron-org/neutron/v6/app/params"
)

func StakeWithDrop(ctx sdk.Context, sk stakingkeeper.Keeper, bk bankkeeper.Keeper, wk *wasmkeeper.Keeper) error {
	daoBalanceBefore, err := bk.Balance(ctx, &types2.QueryBalanceRequest{
		Address: MainDAOContractAddress,
		Denom:   appparams.DefaultDenom,
	})
	if err != nil {
		return err
	}

	// delegate half, and check, if success, then delegate reminder
	// if drop delegation fails, then delegate with native staking module
	halfDelegation := math.NewInt(StakeWithDropAmount).QuoRaw(2)
	err = DropDelegate(ctx, wk, halfDelegation)
	if err != nil {
		// ignore the error, because we have fallback logic
		ctx.Logger().Error("Drop delegation failed", "error", err)
	}

	dropAddress, err := sdk.AccAddressFromBech32(DropCoreContractAddress)
	if err != nil {
		return fmt.Errorf("failed to parse DropDelegateContract contract address: %w", err)
	}
	// check delegations, they are really exist and assets do not stuck on the drop contract
	delegations, err := sk.GetAllDelegatorDelegations(ctx, dropAddress)
	if err != nil {
		return err
	}

	delegatedByDropShares := math.LegacyZeroDec()
	for _, d := range delegations {
		delegatedByDropShares = delegatedByDropShares.Add(d.Shares)
	}

	toDelegateReminder := math.NewInt(StakeWithDropAmount).Sub(halfDelegation)
	// In general shares(delegatedByDropShares) and tokens(halfDelegation) have a conversion rate that depends on the validator’s prior slashes.
	// However, in this specific case, validators are newly created in the same block, which means
	// they have not been slashed yet. This ensures a 1:1 exchange rate between shares and tokens
	// at this stage, making the **direct comparison valid**.
	ctx.Logger().Info("drop delegated", "wanted", halfDelegation, "got", delegatedByDropShares)
	if delegatedByDropShares.GTE(math.LegacyNewDecFromInt(halfDelegation)) {
		// drop delegation finished, delegate remainder
		ctx.Logger().Info("delegating reminder with drop", "amount", toDelegateReminder)
		err = DropDelegate(ctx, wk, toDelegateReminder)
		if err != nil {
			return err
		}
	} else {
		daoBalanceAfter, err := bk.Balance(ctx, &types2.QueryBalanceRequest{
			Address: MainDAOContractAddress,
			Denom:   appparams.DefaultDenom,
		})
		if err != nil {
			return err
		}
		if daoBalanceAfter.Balance.Amount.Equal(daoBalanceBefore.Balance.Amount) {
			// dao completely failed to delegate with drop
			toDelegateReminder = math.NewInt(StakeWithDropAmount)
		}
		ctx.Logger().Info("delegating reminder native way", "amount", toDelegateReminder)
		// fallback to native staking
		err = NativeDelegation(ctx, sk, toDelegateReminder)
		if err != nil {
			return err
		}
	}

	return nil
}

func NativeDelegation(ctx sdk.Context, sk stakingkeeper.Keeper, amount math.Int) error {
	srv := stakingkeeper.NewMsgServerImpl(&sk)
	validators, err := sk.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	validatorsDelegateTo := []types.Validator{}
	for _, validator := range validators {
		// all new staking validator at this stage in types.Unbonded status
		// where all ICS migrated in types.Bonded
		// see DeICS function
		if validator.Status == types.Unbonded {
			validatorsDelegateTo = append(validatorsDelegateTo, validator)
		}
	}
	// distribute stake equally
	// for example, we have 11 to stake and 3 vals
	// amountPerValidator = 3
	// reminder = 2
	// toValidate 3+1, 3+1, 3
	amountPerValidator := amount.QuoRaw(int64(len(validatorsDelegateTo)))
	reminder := amount.ModRaw(int64(len(validatorsDelegateTo)))
	for i, validator := range validatorsDelegateTo {
		toValidate := amountPerValidator
		if int64(i) < reminder.Int64() {
			toValidate = toValidate.AddRaw(1)
		}
		_, err = srv.Delegate(ctx, &types.MsgDelegate{
			DelegatorAddress: MainDAOContractAddress,
			ValidatorAddress: validator.OperatorAddress,
			Amount:           sdk.NewCoin(appparams.DefaultDenom, toValidate),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// DropDelegate performs delegation with DROP protocol
//
//	General Delegation Mechanism via Drop Protocol:
//	1. Execute **bond**.
//	2. Execute **tick**.
//
// However, during testing, we discovered that **tick** does not always process the delegation queue.
//
// After consulting with the Drop team, we decided to slightly adjust the algorithm:
//  1. Execute **tick**.
//  2. Execute **bond**.
//  3. Execute **tick** again.
//
// This adjustment aims to ensure that the protocol successfully delegates coins during the upgrade process.
func DropDelegate(ctx sdk.Context, wk *wasmkeeper.Keeper, amount math.Int) error {
	wasmSrv := wasmkeeper.NewMsgServerImpl(wk)

	msgTick, err := json.Marshal(DropExecuteMsg{
		Tick: &struct{}{},
	})
	if err != nil {
		return err
	}

	// see description above why we need the call
	_, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: DropCoreContractAddress,
		Msg:      msgTick,
		Funds:    nil,
	})
	if err != nil {
		return err
	}

	msgDelegate, err := json.Marshal(DropExecuteMsg{
		Bond: &struct{}{},
	})
	if err != nil {
		return err
	}

	DropDelegateCoins := sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, amount))

	// Start drop delegation
	_, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: DropCoreContractAddress,
		Msg:      msgDelegate,
		Funds:    DropDelegateCoins,
	})
	if err != nil {
		return err
	}

	// execute delayed drop delegation
	_, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: DropCoreContractAddress,
		Msg:      msgTick,
		Funds:    nil,
	})
	if err != nil {
		return err
	}

	return nil
}

type DropExecuteMsg struct {
	Bond *struct{} `json:"bond,omitempty"`
	Tick *struct{} `json:"tick,omitempty"`
}
