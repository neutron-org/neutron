package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// GenesisState is the genesis state of the blockchain represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	// This ugly hack is required to alter globalfee module genesis state
	// because in current chain implementation staking module is absent which is required by globalfee module
	// and we can't use default genesis state for globalfee module.
	// If we will not alter globalfee module genesis state, then we will get panic during tests run.

	genesisState := ModuleBasics.DefaultGenesis(cdc)
	minGasPrices := json.RawMessage(`{"params":{"minimum_gas_prices":[{"denom": "untrn", "amount": "0"}]}}`)
	genesisState["globalfee"] = minGasPrices

	return genesisState
}
