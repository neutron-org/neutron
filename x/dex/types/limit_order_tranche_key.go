package types

import (
	math_utils "github.com/neutron-org/neutron/v6/utils/math"
)

var _ TickLiquidityKey = (*LimitOrderTrancheKey)(nil)

func (p LimitOrderTrancheKey) KeyMarshal() []byte {
	var key []byte

	pairKeyBytes := []byte(p.TradePairId.MustPairID().CanonicalString())
	key = append(key, pairKeyBytes...)
	key = append(key, []byte("/")...)

	makerDenomBytes := []byte(p.TradePairId.MakerDenom)
	key = append(key, makerDenomBytes...)
	key = append(key, []byte("/")...)

	tickIndexBytes := TickIndexToBytes(p.TickIndexTakerToMaker)
	key = append(key, tickIndexBytes...)
	key = append(key, []byte("/")...)

	liquidityTypeBytes := []byte(LiquidityTypeLimitOrder)
	key = append(key, liquidityTypeBytes...)
	key = append(key, []byte("/")...)

	key = append(key, []byte(p.TrancheKey)...)
	key = append(key, []byte("/")...)

	return key
}

func (p LimitOrderTrancheKey) Price() (priceTakerToMaker math_utils.PrecDec, err error) {
	return CalcPrice(p.TickIndexTakerToMaker)
}

func (p LimitOrderTrancheKey) MustPrice() (priceTakerToMaker math_utils.PrecDec) {
	price, err := p.Price()
	if err != nil {
		panic(err)
	}
	return price
}
