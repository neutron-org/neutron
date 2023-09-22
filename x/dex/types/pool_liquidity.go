package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/utils/math"
)

type PoolLiquidity struct {
	TradePairID *TradePairID
	Pool        *Pool
}

func (pl *PoolLiquidity) Swap(
	maxAmountTakerDenomIn sdk.Int,
	maxAmountMakerDenomOut *sdk.Int,
) (inAmount, outAmount sdk.Int) {
	return pl.Pool.Swap(
		pl.TradePairID,
		maxAmountTakerDenomIn,
		maxAmountMakerDenomOut,
	)
}

func (pl *PoolLiquidity) Price() math_utils.PrecDec {
	return pl.Pool.Price(pl.TradePairID)
}
