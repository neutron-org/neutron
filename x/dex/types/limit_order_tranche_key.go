package types

import (
	"errors"
	"strings"

	math_utils "github.com/neutron-org/neutron/utils/math"
)

var _ TickLiquidityKey = (*LimitOrderTrancheKey)(nil)

func (p LimitOrderTrancheKey) KeyMarshal() []byte {
	var key []byte

	pairKeyBytes := []byte(p.TradePairID.MustPairID().CanonicalString())
	key = append(key, pairKeyBytes...)
	key = append(key, []byte("/")...)

	makerDenomBytes := []byte(p.TradePairID.MakerDenom)
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

func (p LimitOrderTrancheKey) KeyUnmarshal(bz []byte) error {
	split := strings.Split(string(bz), "/")

	if len(split) != 5 {
		return errors.New("invalid input length")
	}

	pairKey, err := NewPairIDFromCanonicalString(split[0])
	if err != nil {
		return err
	}
	p.TradePairID = pairKey.MustTradePairIDFromMaker(split[1])

	tickIndex, err := BytesToTickIndex([]byte(split[2]))
	if err != nil {
		return err
	}
	p.TickIndexTakerToMaker = tickIndex

	if split[3] != LiquidityTypeLimitOrder {
		return errors.New("unexpected liquidity type")
	}

	p.TrancheKey = split[4]

	return nil
}

func (p LimitOrderTrancheKey) PriceTakerToMaker() (priceTakerToMaker math_utils.PrecDec, err error) {
	return CalcPrice(p.TickIndexTakerToMaker)
}

func (p LimitOrderTrancheKey) MustPriceTakerToMaker() (priceTakerToMaker math_utils.PrecDec) {
	price, err := p.PriceTakerToMaker()
	if err != nil {
		panic(err)
	}
	return price
}
