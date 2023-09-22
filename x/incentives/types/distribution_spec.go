package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type DistributionSpec map[string]sdk.Coins

func (spec *DistributionSpec) Add(other DistributionSpec) DistributionSpec {
	result := *spec
	for k, v := range other {
		if vv, ok := result[k]; ok {
			result[k] = vv.Add(v...)
		} else {
			result[k] = v
		}
	}
	return result
}

func (spec DistributionSpec) GetTotal() sdk.Coins {
	coins := sdk.Coins{}
	for _, v := range spec {
		coins = coins.Add(v...)
	}
	return coins
}
