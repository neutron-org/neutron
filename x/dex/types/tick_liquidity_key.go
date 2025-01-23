package types

import math_utils "github.com/neutron-org/neutron/v5/utils/math"

type TickLiquidityKey interface {
	KeyMarshal() []byte
	Price() (priceTakerToMaker math_utils.PrecDec, err error)
}
