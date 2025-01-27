package keeper

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks is a wrapper struct for hooks used by the harpoon keeper.
// These hooks are invoked by the staking module.
type Hooks struct {
	k Keeper
}

// AfterValidatorBonded calls the sudo method on contracts subscribed to the AfterValidatorBonded hook.
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorBonded{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorBonded, message)
}

// AfterValidatorRemoved calls the sudo method on the contracts subscribed to the AfterValidatorRemoved hook
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorRemoved{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorRemoved, message)
}

// AfterValidatorCreated calls the sudo method on the contracts subscribed to the AfterValidatorCreated hook
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorCreated{
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorCreated, message)
}

// AfterValidatorBeginUnbonding calls the sudo method on the contracts subscribed to the AfterValidatorBeginUnbonding hook
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorBeginUnbonding{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorBeginUnbonding, message)
}

// BeforeValidatorModified calls the sudo method on the contracts subscribed to the BeforeValidatorModified hook
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeValidatorModified{
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeValidatorModified, message)
}

// BeforeDelegationCreated calls the sudo method on the contracts subscribed to the BeforeDelegationCreated hook
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationCreated{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationCreated, message)
}

// BeforeDelegationSharesModified calls the sudo method on the contracts subscribed to the BeforeDelegationSharesModified hook
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationSharesModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationSharesModified, message)
}

// BeforeDelegationRemoved calls the sudo method on the contracts subscribed to the BeforeDelegationRemoved hook
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationRemoved{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationRemoved, message)
}

// AfterDelegationModified calls the sudo method on the contracts subscribed to the AfterDelegationModified hook
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterDelegationModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterDelegationModified, message)
}

// BeforeValidatorSlashed calls the sudo method on the contracts subscribed to the BeforeValidatorSlashed hook
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	message := types.SudoBeforeValidatorSlashed{
		ValAddr:  valAddr,
		Fraction: fraction,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeValidatorSlashed, message)
}

// AfterUnbondingInitiated calls the sudo method on the contracts subscribed to the AfterUnbondingInitiated hook
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	message := types.SudoAfterUnbondingInitiated{
		Id: id,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterUnbondingInitiated, message)
}
