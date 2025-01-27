package types

import "cosmossdk.io/math"

// This file defines sudo call message structures for Wasm contracts for each hook type.

type SudoAfterValidatorBonded struct {
	ConsAddr []byte `json:"cons_addr"`
	ValAddr  []byte `json:"val_addr"`
}

type SudoAfterValidatorRemoved struct {
	ConsAddr []byte `json:"cons_addr"`
	ValAddr  []byte `json:"val_addr"`
}

type SudoAfterValidatorCreated struct {
	ValAddr []byte `json:"val_addr"`
}

type SudoAfterValidatorBeginUnbonding struct {
	ConsAddr []byte `json:"cons_addr"`
	ValAddr  []byte `json:"val_addr"`
}

type SudoBeforeValidatorModified struct {
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeDelegationCreated struct {
	DelAddr []byte `json:"del_addr"`
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeDelegationSharesModified struct {
	DelAddr []byte `json:"del_addr"`
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeDelegationRemoved struct {
	DelAddr []byte `json:"del_addr"`
	ValAddr []byte `json:"val_addr"`
}

type SudoAfterDelegationModified struct {
	DelAddr []byte `json:"del_addr"`
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeValidatorSlashed struct {
	ValAddr  []byte         `json:"val_addr"`
	Fraction math.LegacyDec `json:"fraction"`
}

type SudoAfterUnbondingInitiated struct {
	Id uint64 `json:"id"`
}
