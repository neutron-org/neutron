package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmKeeper defines the expected interface for the Wasm module.
type WasmKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
	Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
