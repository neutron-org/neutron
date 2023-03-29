package types

import (
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContractOpsKeeper defines the expected contract ops keeper used for simulations (noalias)
type ContractOpsKeeper interface {
	types.ContractOpsKeeper
	// Methods imported from account should be defined here
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	// Methods imported from account should be defined here
}
