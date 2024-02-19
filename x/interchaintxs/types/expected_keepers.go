package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	feerefundertypes "github.com/neutron-org/neutron/v2/x/feerefunder/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

type ICAControllerKeeper interface {
	GetActiveChannelID(ctx sdk.Context, connectionID, portID string) (string, bool)
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
	RegisterInterchainAccount(ctx sdk.Context, connectionID, owner, version string) error
	SendTx(ctx sdk.Context, chanCap *capabilitytypes.Capability, connectionID, portID string, icaPacketData icatypes.InterchainAccountPacketData, timeoutTimestamp uint64) (uint64, error)
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
	GetConnection(ctx sdk.Context, connectionID string) (ibcexported.ConnectionI, error)
}
