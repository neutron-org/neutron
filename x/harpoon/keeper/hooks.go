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
	// TODO: find a way to unify this, since it's gonna be the same for all hooks
	// TODO: more efficient system of storage? since we need to filter out needed subscriptions often
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorBonded)
	message := types.SudoAfterValidatorBonded{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	h.k.CallSudoForSubscriptions(ctx, subscriptions, message)
	return nil
}

// AfterValidatorRemoved deletes the address-pubkey relation when a validator is removed,
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	subscriptions := h.k.GetSubscribedAddressesForHookType(ctx, types.HookType_AfterValidatorRemoved)
	message := types.SudoAfterValidatorRemoved{
		ConsAddr: consAddr,
		ValAddr:  valAddr,
	}
	h.k.CallSudoForSubscriptions(ctx, subscriptions, message)
	return nil
}

// AfterValidatorCreated adds the address-pubkey relation when a validator is created.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	return nil
}

// TODO: what is this and is it in the current main of cosmos-sdk? why removed?
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, _ uint64) error {
	return nil
}
