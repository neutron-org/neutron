package upgrades

import (
	"context"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type WasmMsgServer interface {
	MigrateContract(context.Context, *wasmtypes.MsgMigrateContract) (*wasmtypes.MsgMigrateContractResponse, error)
	// Methods imported from account should be defined here
}
