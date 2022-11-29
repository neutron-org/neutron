package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// ContractManagerKeeper defines the expected interface needed to add ack information about sudo failure.
type ContractManagerKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	AddContractFailure(ctx sdk.Context, address string, ackID uint64, ackType string)
	SudoResponse(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) ([]byte, error)
	SudoError(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, details string) ([]byte, error)
	SudoTimeout(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) ([]byte, error)
}
