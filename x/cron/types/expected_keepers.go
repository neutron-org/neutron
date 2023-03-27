package types

import (
	"github.com/CosmWasm/wasmd/x/wasm/types"
)

// ContractOpsKeeper defines the expected account keeper used for simulations (noalias)
type ContractOpsKeeper interface {
	types.ContractOpsKeeper
	// Methods imported from account should be defined here
}
