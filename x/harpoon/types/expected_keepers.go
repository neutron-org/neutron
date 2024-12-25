package types

import (
	"context"
	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	//GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation // TODO: remove if unused
	GetModuleAddress(moduleName string) sdk.AccAddress
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}

type WasmMsgServer interface {
	ExecuteContract(context.Context, *wasmtypes.MsgExecuteContract) (*wasmtypes.MsgExecuteContractResponse, error)
	// Methods imported from account should be defined here
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error                           // Must be called when a validator is created
	BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error                         // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator is deleted

	AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator begins unbonding

	BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error        // Must be called when a delegation is created
	BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error // Must be called when a delegation's shares are modified
	BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error        // Must be called when a delegation is removed
	AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error
	BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error

	//AfterUnbondingInitiated//TODO: need it?
}
