package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Gauges []*Gauge

func (g Gauges) GetCoinsDistributed() sdk.Coins {
	result := sdk.Coins{}
	for _, gauge := range g {
		result = result.Add(gauge.DistributedCoins...)
	}

	return result
}

// getToDistributeCoinsFromGauges returns coins that have not been distributed yet from the provided gauges
func (g Gauges) GetCoinsRemaining() sdk.Coins {
	result := sdk.Coins{}

	for _, gauge := range g {
		result = result.Add(gauge.CoinsRemaining()...)
	}
	return result
}
