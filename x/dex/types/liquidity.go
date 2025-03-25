package types

import (
	"cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
)

type Liquidity interface {
	Swap(maxAmountTakerIn math.Int, maxAmountMakerOut *math.Int) (inAmount, outAmount math.Int)
	Price() math_utils.PrecDec
}
