package keeper

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.StakingHooks = Hooks{}

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

// AfterValidatorBonded updates the signing info start height or create a new signing info
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorBonded{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorBonded, message)
}

// AfterValidatorRemoved deletes the address-pubkey relation when a validator is removed,
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorRemoved{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorRemoved, message)
}

// AfterValidatorCreated adds the address-pubkey relation when a validator is created.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorCreated{
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorCreated, message)
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterValidatorBeginUnbonding{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterValidatorBeginUnbonding, message)
}

func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeValidatorModified{
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeValidatorModified, message)
}

func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationCreated{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationCreated, message)
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationSharesModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationSharesModified, message)
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoBeforeDelegationRemoved{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeDelegationRemoved, message)
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.SudoAfterDelegationModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterDelegationModified, message)
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	message := types.SudoBeforeValidatorSlashed{
		ValAddr:  valAddr,
		Fraction: fraction,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_BeforeValidatorSlashed, message)
}

func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	message := types.SudoAfterUnbondingInitiated{
		Id: id,
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HookType_AfterUnbondingInitiated, message)
}
