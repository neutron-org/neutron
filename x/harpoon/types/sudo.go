package types

// TODO: can these types be autogeneratable? this will help to avoid duplication for neutron-std library
type SudoAfterValidatorBonded struct {
	ConsAddr []byte `json:"cons_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr  []byte `json:"val_addr"`
}

type SudoAfterValidatorRemoved struct {
	ConsAddr []byte `json:"cons_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr  []byte `json:"val_addr"`
}

type SudoAfterValidatorCreated struct {
	ValAddr []byte `json:"val_addr"`
}

type SudoAfterValidatorBeginUnbonding struct {
	ConsAddr []byte `json:"cons_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr  []byte `json:"val_addr"`
}

type SudoBeforeValidatorModified struct {
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeDelegationCreated struct {
	DelAddr []byte `json:"del_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeDelegationSharesModified struct {
	DelAddr []byte `json:"del_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeDelegationRemoved struct {
	DelAddr []byte `json:"del_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr []byte `json:"val_addr"`
}

type SudoAfterDelegationModified struct {
	DelAddr []byte `json:"del_addr"` // TODO: can we make type a string or an address of some kind?
	ValAddr []byte `json:"val_addr"`
}

type SudoBeforeValidatorSlashed struct {
	DelAddr  []byte `json:"del_addr"` // TODO: can we make type a string or an address of some kind?
	Fraction string `json:"fraction"` // TODO: how to serialize LegacyDec?
}

// type SudoAfterUnbondingInitiated struct {
//
//}
