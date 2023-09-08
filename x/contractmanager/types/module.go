package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

type ContractManagerWrapper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	AddContractFailure(ctx sdk.Context, packet *channeltypes.Packet, address, ackType string, ack *channeltypes.Acknowledgement)
	SudoResponse(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, ack channeltypes.Acknowledgement) ([]byte, error)
	SudoError(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, ack channeltypes.Acknowledgement) ([]byte, error)
	SudoTimeout(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) ([]byte, error)
	SudoOnChanOpenAck(ctx sdk.Context, contractAddress sdk.AccAddress, details OpenAckDetails) ([]byte, error)
	GetParams(ctx sdk.Context) (params Params)
}
