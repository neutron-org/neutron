package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
)

// ContractManagerKeeper defines the expected interface needed to add ack information about sudo failure.
type ContractManagerKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	AddContractFailure(ctx sdk.Context, channelID string, address string, ackID uint64, ackType string)
	SudoResponse(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) ([]byte, error)
	SudoError(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, details string) ([]byte, error)
	SudoTimeout(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) ([]byte, error)
}

type FeeRefunderKeeper interface {
	LockFees(ctx sdk.Context, payer sdk.AccAddress, packetID feerefundertypes.PacketID, fee feerefundertypes.Fee) error
	DistributeAcknowledgementFee(ctx sdk.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
	DistributeTimeoutFee(ctx sdk.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
}

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}
