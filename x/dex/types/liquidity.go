package types

import (
	math_utils "github.com/neutron-org/neutron/v8/utils/math"
)

type Liquidity interface {
	Swap(maxAmountTakerIn math_utils.PrecDec, maxAmountMakerOut *math_utils.PrecDec) (inAmount, outAmount math_utils.PrecDec)
	Price() math_utils.PrecDec
}
