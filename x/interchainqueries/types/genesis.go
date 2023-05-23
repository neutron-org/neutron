package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}
	seenIDs := map[uint64]bool{}

	for _, val := range gs.GetRegisteredQueries() {
		if seenIDs[val.Id] {
			return sdkerrors.Wrapf(ErrInvalidQueryID, "duplicate query id: %d", val.Id)
		}
		seenIDs[val.Id] = true

		_, err = sdk.AccAddressFromBech32(val.Owner)
		if err != nil {
			return sdkerrors.Wrapf(err, "Invalid owner address (%s)", err)
		}

		if val.QueryType == "" {
			return sdkerrors.Wrapf(ErrEmptyQueryTypeGenesis, "Query type is empty, id: %d", val.Id)
		}

		if val.QueryType == "tx" {
			if err := ValidateTransactionsFilter(val.TransactionsFilter); err != nil {
				return sdkerrors.Wrap(ErrInvalidTransactionsFilter, err.Error())
			}
		}
		if val.QueryType == "kv" {
			if len(val.Keys) == 0 {
				return sdkerrors.Wrap(ErrEmptyKeys, "keys cannot be empty")
			}
			if err := validateKeys(val.GetKeys()); err != nil {
				return err
			}
		}
	}
	return nil
}
