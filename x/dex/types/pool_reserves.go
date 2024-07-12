package types

import (
	"cosmossdk.io/math"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
)

func (p PoolReserves) HasToken() bool {
	return p.ReservesMakerDenom.GT(math.ZeroInt())
}

func NewPoolReservesFromCounterpart(
	counterpart *PoolReserves,
) *PoolReserves {
	thisID := counterpart.Key.Counterpart()
	//TODO: Is this safe?
	makerPrice := MustCalcPrice(thisID.TickIndexTakerToMaker)
	return &PoolReserves{
		Key:                thisID,
		ReservesMakerDenom: math.ZeroInt(),
		MakerPrice:         makerPrice,
		PriceTakerToMaker:  math_utils.OnePrecDec().Quo(makerPrice),
	}
}

func NewPoolReserves(
	poolReservesID *PoolReservesKey,
) (*PoolReserves, error) {
	makerPrice, err := poolReservesID.Price()
	if err != nil {
		return nil, err
	}

	return &PoolReserves{
		Key:                poolReservesID,
		ReservesMakerDenom: math.ZeroInt(),
		MakerPrice:         makerPrice,
		PriceTakerToMaker:  math_utils.OnePrecDec().Quo(makerPrice),
	}, nil
}

func MustNewPoolReserves(
	poolReservesID *PoolReservesKey,
) *PoolReserves {
	poolReserves, err := NewPoolReserves(poolReservesID)
	if err != nil {
		panic(err)
	}
	return poolReserves
}
