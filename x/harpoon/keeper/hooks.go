package keeper

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks wrapper struct for harpoon keeper.
// These hooks are called by the staking module.
type Hooks struct {
	k Keeper
}

// AfterValidatorBonded calls sudo method on the contracts subscribed to AfterValidatorBonded hook
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorBonded{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorBonded, message)
}

// AfterValidatorRemoved calls sudo method on the contracts subscribed to AfterValidatorRemoved hook
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorRemoved{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorRemoved, message)
}

// AfterValidatorCreated calls sudo method on the contracts subscribed to AfterValidatorCreated hook
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorCreated{
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorCreated, message)
}

// AfterValidatorBeginUnbonding calls sudo method on the contracts subscribed to AfterValidatorBeginUnbonding hook
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorBeginUnbonding{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorBeginUnbonding, message)
}

// BeforeValidatorModified calls sudo method on the contracts subscribed to BeforeValidatorModified hook
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeValidatorModified{
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeValidatorModified, message)
}

// BeforeDelegationCreated calls sudo method on the contracts subscribed to BeforeDelegationCreated hook
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationCreated{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationCreated, message)
}

// BeforeDelegationSharesModified calls sudo method on the contracts subscribed to BeforeDelegationSharesModified hook
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationSharesModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationSharesModified, message)
}

// BeforeDelegationRemoved calls sudo method on the contracts subscribed to BeforeDelegationRemoved hook
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationRemoved{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationRemoved, message)
}

// AfterDelegationModified calls sudo method on the contracts subscribed to AfterDelegationModified hook
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterDelegationModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterDelegationModified, message)
}

// BeforeValidatorSlashed calls sudo method on the contracts subscribed to BeforeValidatorSlashed hook
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	message := types.SudoBeforeValidatorSlashed{
		ValAddr:  valAddr,
		Fraction: fraction,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeValidatorSlashed, message)
}

// AfterUnbondingInitiated calls sudo method on the contracts subscribed to AfterUnbondingInitiated hook
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	message := types.SudoAfterUnbondingInitiated{
		Id: id,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterUnbondingInitiated, message)
}
