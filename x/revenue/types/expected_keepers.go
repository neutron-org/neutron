package types

import (
	"context"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"
	"math/big"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
)

// VoteAggregator defines the expected interface for aggregating oracle votes into a set of prices.
type VoteAggregator interface {
	// GetPriceForValidator gets the prices reported by a given validator. This method depends
	// on the prices from the latest set of aggregated votes.
	GetPriceForValidator(validator sdktypes.ConsAddress) map[slinkytypes.CurrencyPair]*big.Int
}

// StakingKeeper defines the expected interface for getting validators by consensus address.
type StakingKeeper interface {
	// GetValidatorByConsAddr gets a single validator by consensus address
	GetValidatorByConsAddr(ctx context.Context, consAddr sdktypes.ConsAddress) (validator stakingtypes.Validator, err error)
}

// BankKeeper defines the expected interface needed to send coins from one account to another.
type BankKeeper interface {
	// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdktypes.AccAddress, amt sdktypes.Coins) error
}

// OracleKeeper defines the expected interfaces getting oracle price for a given asset pair.
type OracleKeeper interface {
	// GetPriceForCurrencyPair gets a price for a given asset pair
	GetPriceForCurrencyPair(ctx sdktypes.Context, cp slinkytypes.CurrencyPair) (oracletypes.QuotePrice, error)

	// GetDecimalsForCurrencyPair gets decimals for a given pair
	GetDecimalsForCurrencyPair(ctx sdktypes.Context, cp slinkytypes.CurrencyPair) (decimals uint64, err error)
}
