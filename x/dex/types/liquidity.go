package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/utils/math"
)

type Liquidity interface {
	Swap(maxAmountTakerIn sdk.Int, maxAmountMakerOut *sdk.Int) (inAmount, outAmount sdk.Int)
	Price() math_utils.PrecDec
}
