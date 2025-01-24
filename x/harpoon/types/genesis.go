package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		HookSubscriptions: make([]HookSubscriptions, 0),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, item := range gs.HookSubscriptions {
		if _, ok := HookType_name[int32(item.HookType)]; !ok {
			return fmt.Errorf("invalid hook type: %d", int32(item.HookType))
		}
		for _, addr := range item.ContractAddresses {
			if _, err := sdk.AccAddressFromBech32(addr); err != nil {
				return fmt.Errorf("invalid contract_address=%s in genesis state: %v", addr, err)
			}

		}
	}

	return nil
}
