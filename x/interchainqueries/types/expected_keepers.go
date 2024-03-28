package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	// Methods imported from bank should be defined here
}

type ContractManagerKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
	SudoKVQueryResult(ctx context.Context, contractAddress sdk.AccAddress, queryID uint64) ([]byte, error)
	SudoTxQueryResult(ctx context.Context, contractAddress sdk.AccAddress, queryID uint64, height ibcclienttypes.Height, data []byte) ([]byte, error)
}
