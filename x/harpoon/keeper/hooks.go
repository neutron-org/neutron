package keeper

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ stakingtypes.StakingHooks                                      = Hooks{}
	_ stakingtypes.StakingHooksBeforeValidatorSlashedHasTokensToBurn = Hooks{}
)

// Hooks is a wrapper struct for hooks used by the harpoon keeper.
// These hooks are invoked by the staking module.
type Hooks struct {
	k *Keeper
}

// AfterValidatorBonded calls the sudo method on contracts subscribed to the AfterValidatorBonded hook.
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.AfterValidatorBondedSudoMsg{
		AfterValidatorBonded: types.AfterValidatorBondedMsg{
			ConsAddr: consAddr.String(),
			ValAddr:  valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED, message)
}

// AfterValidatorRemoved calls the sudo method on the contracts subscribed to the AfterValidatorRemoved hook
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.AfterValidatorRemovedSudoMsg{
		AfterValidatorRemoved: types.AfterValidatorRemovedMsg{
			ConsAddr: consAddr.String(),
			ValAddr:  valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_REMOVED, message)
}

// AfterValidatorCreated calls the sudo method on the contracts subscribed to the AfterValidatorCreated hook
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.AfterValidatorCreatedSudoMsg{
		AfterValidatorCreated: types.AfterValidatorCreatedMsg{
			ValAddr: valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, message)
}

// AfterValidatorBeginUnbonding calls the sudo method on the contracts subscribed to the AfterValidatorBeginUnbonding hook
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	message := types.AfterValidatorBeginUnbondingSudoMsg{
		AfterValidatorBeginUnbonding: types.AfterValidatorBeginUnbondingMsg{
			ConsAddr: consAddr.String(),
			ValAddr:  valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_VALIDATOR_BEGIN_UNBONDING, message)
}

// BeforeValidatorModified calls the sudo method on the contracts subscribed to the BeforeValidatorModified hook
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	message := types.BeforeValidatorModifiedSudoMsg{
		BeforeValidatorModified: types.BeforeValidatorModifiedMsg{
			ValAddr: valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_BEFORE_VALIDATOR_MODIFIED, message)
}

// BeforeDelegationCreated calls the sudo method on the contracts subscribed to the BeforeDelegationCreated hook
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.BeforeDelegationCreatedSudoMsg{
		BeforeDelegationCreated: types.BeforeDelegationCreatedMsg{
			DelAddr: delAddr.String(),
			ValAddr: valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_BEFORE_DELEGATION_CREATED, message)
}

// BeforeDelegationSharesModified calls the sudo method on the contracts subscribed to the BeforeDelegationSharesModified hook
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.BeforeDelegationSharesModifiedSudoMsg{
		BeforeDelegationSharesModified: types.BeforeDelegationSharesModifiedMsg{
			DelAddr: delAddr.String(),
			ValAddr: valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_BEFORE_DELEGATION_SHARES_MODIFIED, message)
}

// BeforeDelegationRemoved calls the sudo method on the contracts subscribed to the BeforeDelegationRemoved hook
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.BeforeDelegationRemovedSudoMsg{
		BeforeDelegationRemoved: types.BeforeDelegationRemovedMsg{
			DelAddr: delAddr.String(),
			ValAddr: valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_BEFORE_DELEGATION_REMOVED, message)
}

// AfterDelegationModified calls the sudo method on the contracts subscribed to the AfterDelegationModified hook
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	message := types.AfterDelegationModifiedSudoMsg{
		AfterDelegationModified: types.AfterDelegationModifiedMsg{
			DelAddr: delAddr.String(),
			ValAddr: valAddr.String(),
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED, message)
}

// BeforeValidatorSlashed is not implemented because BeforeValidatorSlashedWithTokensToBurn will be called instead
func (h Hooks) BeforeValidatorSlashed(_ context.Context, _ sdk.ValAddress, _ sdkmath.LegacyDec) error {
	panic("BeforeValidatorSlashed shouldn't ever be called for neutron harpoon hooks since it has BeforeValidatorSlashedWithTokensToBurn hook")
}

// BeforeValidatorSlashedWithTokensToBurn calls the sudo method on the contracts subscribed to the BeforeValidatorSlashedWithTokensToBurn hook
// It's same as BeforeValidatorSlashed but with tokensToBurn argument. Made this way for compatibility purposes.
func (h Hooks) BeforeValidatorSlashedWithTokensToBurn(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec, tokensToBurn sdkmath.Int) error {
	message := types.BeforeValidatorSlashedSudoMsg{
		BeforeValidatorSlashed: types.BeforeValidatorSlashedMsg{
			ValAddr:      valAddr.String(),
			Fraction:     fraction,
			TokensToBurn: tokensToBurn,
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_BEFORE_VALIDATOR_SLASHED, message)
}

// AfterUnbondingInitiated calls the sudo method on the contracts subscribed to the AfterUnbondingInitiated hook
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	message := types.AfterUnbondingInitiatedSudoMsg{
		AfterUnbondingInitiated: types.AfterUnbondingInitiatedMsg{
			ID: id,
		},
	}
	return h.k.CallSudoForSubscriptionType(ctx, types.HOOK_TYPE_AFTER_UNBONDING_INITIATED, message)
}
