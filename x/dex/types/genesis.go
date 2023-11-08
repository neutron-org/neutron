package types

import (
	"fmt"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		LimitOrderTrancheUserList:     []*LimitOrderTrancheUser{},
		TickLiquidityList:             []*TickLiquidity{},
		InactiveLimitOrderTrancheList: []*LimitOrderTranche{},
		PoolMetadataList:              []PoolMetadata{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in LimitOrderTrancheUser
	LimitOrderTrancheUserIndexMap := make(map[string]struct{})

	for _, elem := range gs.LimitOrderTrancheUserList {
		index := string(LimitOrderTrancheUserKey(elem.Address, elem.TrancheKey))
		if _, ok := LimitOrderTrancheUserIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for LimitOrderTrancheUser")
		}
		LimitOrderTrancheUserIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in tickLiquidity
	tickLiquidityIndexMap := make(map[string]struct{})

	for _, elem := range gs.TickLiquidityList {
		var index string
		switch liquidity := elem.Liquidity.(type) {
		case *TickLiquidity_PoolReserves:
			index = string(liquidity.PoolReserves.Key.KeyMarshal())
		case *TickLiquidity_LimitOrderTranche:
			index = string(liquidity.LimitOrderTranche.Key.KeyMarshal())
		}
		if _, ok := tickLiquidityIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for tickLiquidity")
		}
		tickLiquidityIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in inactiveLimitOrderTranche
	inactiveLimitOrderTrancheKeyMap := make(map[string]struct{})

	for _, elem := range gs.InactiveLimitOrderTrancheList {
		index := string(elem.Key.KeyMarshal())
		if _, ok := inactiveLimitOrderTrancheKeyMap[index]; ok {
			return fmt.Errorf("duplicated index for inactiveLimitOrderTranche")
		}
		inactiveLimitOrderTrancheKeyMap[index] = struct{}{}
	}
	// Check for duplicated ID in poolMetadata
	poolMetadataIDMap := make(map[uint64]bool)
	poolMetadataCount := gs.GetPoolCount()
	for _, elem := range gs.PoolMetadataList {
		if _, ok := poolMetadataIDMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for poolMetadata")
		}
		if elem.Id >= poolMetadataCount {
			return fmt.Errorf("poolMetadata id should be lower or equal than the last id")
		}
		poolMetadataIDMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
