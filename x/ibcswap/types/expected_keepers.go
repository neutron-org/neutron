package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface that the swap middleware needs in order to facilitate refunds.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
