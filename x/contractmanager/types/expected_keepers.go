package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmKeeper defines the expected interface needed to cam cosmwasm contracts.
type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

type ContractManagerKeeper interface {
	AddContractFailure(ctx sdk.Context, address string, sudoPayload []byte)
	GetParams(ctx sdk.Context) (params Params)
}
