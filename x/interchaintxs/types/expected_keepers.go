package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"

	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

type ContractManagerKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	AddContractFailure(ctx sdk.Context, channelID string, address string, ackID uint64, ackType string)
	SudoResponse(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) ([]byte, error)
	SudoError(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, details string) ([]byte, error)
	SudoTimeout(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) ([]byte, error)
	SudoOnChanOpenAck(ctx sdk.Context, contractAddress sdk.AccAddress, details contractmanagertypes.OpenAckDetails) ([]byte, error)
}

type ICAControllerKeeper interface {
	GetActiveChannelID(ctx sdk.Context, connectionID, portID string) (string, bool)
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
	RegisterInterchainAccount(ctx sdk.Context, connectionID, owner string, version string) error
	SendTx(ctx sdk.Context, chanCap *capabilitytypes.Capability, connectionID, portID string, icaPacketData icatypes.InterchainAccountPacketData, timeoutTimestamp uint64) (uint64, error)
}

type FeeRefunderKeeper interface {
	LockFees(ctx sdk.Context, payer sdk.AccAddress, packetID feerefundertypes.PacketID, fee feerefundertypes.Fee) error
	DistributeAcknowledgementFee(ctx sdk.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
	DistributeTimeoutFee(ctx sdk.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
}

type ScopedKeeper interface {
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
	GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetConnection(ctx sdk.Context, connectionID string) (ibcexported.ConnectionI, error)
}
