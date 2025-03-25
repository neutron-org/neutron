package types

import (
	"cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
)

type PoolLiquidity struct {
	TradePairID *TradePairID
	Pool        *Pool
}

func (pl *PoolLiquidity) Swap(
	maxAmountTakerDenomIn math.Int,
	maxAmountMakerDenomOut *math.Int,
) (inAmount, outAmount math.Int) {
	return pl.Pool.Swap(
		pl.TradePairID,
		maxAmountTakerDenomIn,
		maxAmountMakerDenomOut,
	)
}

func (pl *PoolLiquidity) Price() math_utils.PrecDec {
	return pl.Pool.Price(pl.TradePairID)
}
