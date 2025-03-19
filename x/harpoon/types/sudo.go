package types

import "cosmossdk.io/math"

// This file defines sudo call message structures for Wasm contracts for each hook type.

type AfterValidatorBondedSudoMsg struct {
	AfterValidatorBonded AfterValidatorBondedMsg `json:"after_validator_bonded,omitempty"`
}

type AfterValidatorBondedMsg struct {
	ConsAddr string `json:"cons_addr"`
	ValAddr  string `json:"val_addr"`
}

type AfterValidatorRemovedSudoMsg struct {
	AfterValidatorRemoved AfterValidatorRemovedMsg `json:"after_validator_removed,omitempty"`
}

type AfterValidatorRemovedMsg struct {
	ConsAddr string `json:"cons_addr"`
	ValAddr  string `json:"val_addr"`
}

type AfterValidatorCreatedSudoMsg struct {
	AfterValidatorCreated AfterValidatorCreatedMsg `json:"after_validator_created,omitempty"`
}

type AfterValidatorCreatedMsg struct {
	ValAddr string `json:"val_addr"`
}

type AfterValidatorBeginUnbondingSudoMsg struct {
	AfterValidatorBeginUnbonding AfterValidatorBeginUnbondingMsg `json:"after_validator_begin_unbonding,omitempty"`
}

type AfterValidatorBeginUnbondingMsg struct {
	ConsAddr string `json:"cons_addr"`
	ValAddr  string `json:"val_addr"`
}

type BeforeValidatorModifiedSudoMsg struct {
	BeforeValidatorModified BeforeValidatorModifiedMsg `json:"before_validator_modified,omitempty"`
}

type BeforeValidatorModifiedMsg struct {
	ValAddr string `json:"val_addr"`
}

type BeforeDelegationCreatedSudoMsg struct {
	BeforeDelegationCreated BeforeDelegationCreatedMsg `json:"before_delegation_created,omitempty"`
}

type BeforeDelegationCreatedMsg struct {
	DelAddr string `json:"del_addr"`
	ValAddr string `json:"val_addr"`
}

type BeforeDelegationSharesModifiedSudoMsg struct {
	BeforeDelegationSharesModified BeforeDelegationSharesModifiedMsg `json:"before_delegation_shares_modified,omitempty"`
}

type BeforeDelegationSharesModifiedMsg struct {
	DelAddr string `json:"del_addr"`
	ValAddr string `json:"val_addr"`
}

type BeforeDelegationRemovedSudoMsg struct {
	BeforeDelegationRemoved BeforeDelegationRemovedMsg `json:"before_delegation_removed,omitempty"`
}

type BeforeDelegationRemovedMsg struct {
	DelAddr string `json:"del_addr"`
	ValAddr string `json:"val_addr"`
}

type AfterDelegationModifiedSudoMsg struct {
	AfterDelegationModified AfterDelegationModifiedMsg `json:"after_delegation_modified,omitempty"`
}

type AfterDelegationModifiedMsg struct {
	DelAddr string `json:"del_addr"`
	ValAddr string `json:"val_addr"`
}

type BeforeValidatorSlashedSudoMsg struct {
	BeforeValidatorSlashed BeforeValidatorSlashedMsg `json:"before_validator_slashed,omitempty"`
}

type BeforeValidatorSlashedMsg struct {
	ValAddr      string         `json:"val_addr"`
	Fraction     math.LegacyDec `json:"fraction"`
	TokensToBurn math.Int       `json:"tokens_to_burn"`
}

type AfterUnbondingInitiatedSudoMsg struct {
	AfterUnbondingInitiated AfterUnbondingInitiatedMsg `json:"after_unbonding_initiated,omitempty"`
}

type AfterUnbondingInitiatedMsg struct {
	ID uint64 `json:"id"`
}
