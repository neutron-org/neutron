package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
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
	AddContractFailure(ctx sdk.Context, address string, ackID uint64, ackType string)
	SudoResponse(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) ([]byte, error)
	SudoError(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, details string) ([]byte, error)
	SudoTimeout(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) ([]byte, error)
	SudoOnChanOpenAck(ctx sdk.Context, contractAddress sdk.AccAddress, details contractmanagertypes.OpenAckDetails) ([]byte, error)
}
