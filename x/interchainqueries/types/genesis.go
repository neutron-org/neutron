package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
			return errors.Wrapf(ErrInvalidQueryID, "duplicate query id: %d", val.Id)
		}
		seenIDs[val.Id] = true

		_, err = sdk.AccAddressFromBech32(val.Owner)
		if err != nil {
			return errors.Wrapf(err, "Invalid owner address (%s)", err)
		}

		switch val.QueryType {
		case string(InterchainQueryTypeTX):
			if err := ValidateTransactionsFilter(val.TransactionsFilter, gs.Params.MaxTransactionsFilters); err != nil {
				return errors.Wrap(ErrInvalidTransactionsFilter, err.Error())
			}
		case string(InterchainQueryTypeKV):
			if len(val.Keys) == 0 {
				return errors.Wrap(ErrEmptyKeys, "keys cannot be empty")
			}
			if err := validateKeys(val.GetKeys(), gs.Params.MaxKvQueryKeysCount); err != nil {
				return err
			}
		default:
			return errors.Wrapf(ErrUnexpectedQueryTypeGenesis, "Unexpected query type: %s", val.QueryType)
		}
	}
	return nil
}
