package types

import math_utils "github.com/neutron-org/neutron/v4/utils/math"

type TickLiquidityKey interface {
	KeyMarshal() []byte
	PriceTakerToMaker() (priceTakerToMaker math_utils.PrecDec, err error)
}
