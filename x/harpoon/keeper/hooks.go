package keeper

import (
	"context"
	"cosmossdk.io/errors"
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
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded)
	message := types.SudoAfterValidatorBonded{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_AfterValidatorBonded)
	}
	return nil
}

// AfterValidatorRemoved deletes the address-pubkey relation when a validator is removed,
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorRemoved)
	message := types.SudoAfterValidatorRemoved{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_AfterValidatorRemoved)
	}
	return nil
}

// AfterValidatorCreated adds the address-pubkey relation when a validator is created.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorCreated)
	message := types.SudoAfterValidatorCreated{
		ValAddr: valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_AfterValidatorCreated)
	}
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBeginUnbonding)
	message := types.SudoAfterValidatorBeginUnbonding{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_AfterValidatorBeginUnbonding)
	}
	return nil
}

func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeValidatorModified)
	message := types.SudoBeforeValidatorModified{
		ValAddr: valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_BeforeValidatorModified)
	}
	return nil
}

func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeDelegationCreated)
	message := types.SudoBeforeDelegationCreated{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_BeforeDelegationCreated)
	}
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeDelegationSharesModified)
	message := types.SudoBeforeDelegationSharesModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_BeforeDelegationSharesModified)
	}
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeDelegationRemoved)
	message := types.SudoBeforeDelegationRemoved{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_BeforeDelegationRemoved)
	}
	return nil
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterDelegationModified)
	message := types.SudoAfterDelegationModified{
		DelAddr: delAddr,
		ValAddr: valAddr,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_AfterDelegationModified)
	}
	return nil
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_BeforeValidatorSlashed)
	message := types.SudoBeforeValidatorSlashed{
		ValAddr:  valAddr,
		Fraction: fraction,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_BeforeValidatorSlashed)
	}
	return nil
}

func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterUnbondingInitiated)
	message := types.SudoAfterUnbondingInitiated{
		Id: id,
	}
	if err := h.k.CallSudoForSubscriptions(ctx, subscriptions, message); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", types.HookType_AfterUnbondingInitiated)
	}
	return nil
}
