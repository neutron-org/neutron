package types

import (
	context "context"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper defines the expected interface for getting validators by consensus address.
type StakingKeeper interface {
	// GetValidatorByConsAddr gets a single validator by consensus address
	GetValidatorByConsAddr(ctx context.Context, consAddr sdktypes.ConsAddress) (validator stakingtypes.Validator, err error)
}

// BankKeeper defines the expected interface needed to send coins from one account to another.
type BankKeeper interface {
	// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
	// It will panic if the module account does not exist. An error is returned if
	// the recipient address is black-listed or if sending the tokens fails.
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdktypes.AccAddress, amt sdktypes.Coins) error
	// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount.
	// It will panic if the module account does not exist.
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdktypes.AccAddress, recipientModule string, amt sdktypes.Coins) error
}
