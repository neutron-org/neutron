package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	// Methods imported from bank should be defined here
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	IterateAccountBalances(ctx context.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool)
	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetAccountsBalances(ctx context.Context) []banktypes.Balance
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
