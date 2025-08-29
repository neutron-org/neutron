package types

import (
	"cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
)

func (p PoolReserves) HasToken() bool {
	return p.DecReservesMakerDenom.IsPositive()
}

func NewPoolReservesFromCounterpart(
	counterpart *PoolReserves,
) *PoolReserves {
	thisID := counterpart.Key.Counterpart()
	// Pool tickIndex has already been validated so this will never throw
	makerPrice := MustCalcPrice(thisID.TickIndexTakerToMaker)
	return &PoolReserves{
		Key:                       thisID,
		ReservesMakerDenom:        math.ZeroInt(),
		DecReservesMakerDenom:     math_utils.ZeroPrecDec(),
		MakerPrice:                makerPrice,
		PriceTakerToMaker:         math_utils.OnePrecDec().Quo(makerPrice),
		PriceOppositeTakerToMaker: counterpart.PriceTakerToMaker,
	}
}

func NewPoolReserves(
	poolReservesID *PoolReservesKey,
) (*PoolReserves, error) {
	makerPrice, err := poolReservesID.Price()
	if err != nil {
		return nil, err
	}
	counterpartID := poolReservesID.Counterpart()
	priceOppositeTakerToMaker, err := counterpartID.PriceTakerToMaker()
	if err != nil {
		return nil, err
	}

	return &PoolReserves{
		Key:                       poolReservesID,
		ReservesMakerDenom:        math.ZeroInt(),
		DecReservesMakerDenom:     math_utils.ZeroPrecDec(),
		MakerPrice:                makerPrice,
		PriceTakerToMaker:         math_utils.OnePrecDec().Quo(makerPrice),
		PriceOppositeTakerToMaker: priceOppositeTakerToMaker,
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

func (p *PoolReserves) SetMakerReserves(reserves math_utils.PrecDec) {
	// It should be impossible for reserves to be negative.
	// This check provides an extra safegaurd against possible exploits or bugs.
	if reserves.IsNegative() {
		panic("reserves cannot be negative")
	}
	p.ReservesMakerDenom = reserves.TruncateInt()
	p.DecReservesMakerDenom = reserves
}
