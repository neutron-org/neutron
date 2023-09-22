package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dextypes "github.com/neutron-org/neutron/x/dex/types"
	. "github.com/neutron-org/neutron/x/incentives/types"
)

func TestQueryCondition(t *testing.T) {
	pairID := &dextypes.PairID{
		Token0: "coin1",
		Token1: "coin2",
	}

	tests := []struct {
		name       string
		queryCond  QueryCondition
		poolMetadata dextypes.PoolMetadata
		testResult bool
	}{
		{
			name:       "Matching denom and tick range",
			queryCond:  QueryCondition{PairID: pairID, StartTick: 10, EndTick: 20},
			poolMetadata: dextypes.NewPoolMetadata(pairID, 15, 5, 0),
			testResult: true,
		},
		{
			name:       "Non-matching denom",
			queryCond:  QueryCondition{PairID: pairID, StartTick: 10, EndTick: 20},
			poolMetadata: dextypes.NewPoolMetadata(&dextypes.PairID{Token0: "coin1", Token1: "coin3"}, 15, 5, 0),
			testResult: false,
		},
		{
			name:       "Non-matching tick range",
			queryCond:  QueryCondition{PairID: pairID, StartTick: 30, EndTick: 40},
			poolMetadata: dextypes.NewPoolMetadata(pairID, 15, 6, 0),
			testResult: false,
		},
		{
			name:       "Non-matching tick fee range lower",
			queryCond:  QueryCondition{PairID: pairID, StartTick: 30, EndTick: 40},
			poolMetadata: dextypes.NewPoolMetadata(pairID, 10, 5, 0),
			testResult: false,
		},
		{
			name:       "Non-matching tick fee range upper",
			queryCond:  QueryCondition{PairID: pairID, StartTick: 30, EndTick: 40},
			poolMetadata: dextypes.NewPoolMetadata(pairID, 20, 5, 0),
			testResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.queryCond.Test(tt.poolMetadata)
			assert.Equal(t, tt.testResult, result)
		})
	}
}
