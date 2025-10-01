package types

import (
	math_utils "github.com/neutron-org/neutron/v8/utils/math"
)

type PoolLiquidity struct {
	TradePairID *TradePairID
	Pool        *Pool
}

func (pl *PoolLiquidity) Swap(
	maxAmountTakerDenomIn math_utils.PrecDec,
	maxAmountMakerDenomOut *math_utils.PrecDec,
) (inAmount, outAmount math_utils.PrecDec) {
	return pl.Pool.Swap(
		pl.TradePairID,
		maxAmountTakerDenomIn,
		maxAmountMakerDenomOut,
	)
}

func (pl *PoolLiquidity) Price() math_utils.PrecDec {
	return pl.Pool.Price(pl.TradePairID)
}
