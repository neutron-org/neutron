package types

import (
	"cosmossdk.io/math"
)

func (p PoolReserves) HasToken() bool {
	return p.ReservesMakerDenom.GT(math.ZeroInt())
}

func NewPoolReservesFromCounterpart(
	counterpart *PoolReserves,
) *PoolReserves {
	thisID := counterpart.Key.Counterpart()
	return &PoolReserves{
		Key:                thisID,
		ReservesMakerDenom: math.ZeroInt(),
		//TODO: Is this safe?
		MakerPrice: MustCalcPrice(thisID.TickIndexTakerToMaker),
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
