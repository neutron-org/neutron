package types

import (
	"cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
)

type Liquidity interface {
	Swap(maxAmountTakerIn math.Int, maxAmountMakerOut *math.Int) (inAmount, outAmount math_utils.PrecDec)
	Price() math_utils.PrecDec
}
