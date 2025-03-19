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
	// Check for duplicate hook types
	hookIndexMap := make(map[int32]bool)
	for _, item := range gs.HookSubscriptions {
		if err := ValidateHookType(item.HookType); err != nil {
			return err
		}
		if _, ok := hookIndexMap[int32(item.HookType)]; ok {
			return fmt.Errorf("duplicate hook type: %d", int32(item.HookType))
		}
		hookIndexMap[int32(item.HookType)] = true

		// Check for duplicate contract addresses
		contractAddressIndexMap := make(map[string]bool)
		for _, addr := range item.ContractAddresses {
			if _, err := sdk.AccAddressFromBech32(addr); err != nil {
				return fmt.Errorf("invalid contract_address=%s in genesis state: %v", addr, err)
			}
			if _, ok := contractAddressIndexMap[addr]; ok {
				return fmt.Errorf("duplicate contract address: %s for hook type: %d", addr, int32(item.HookType))
			}
			contractAddressIndexMap[addr] = true
		}
	}

	return nil
}
