package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
)

var _ TickLiquidityKey = (*PoolReservesKey)(nil)

func (p PoolReservesKey) KeyMarshal() []byte {
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

	liquidityTypeBytes := []byte(LiquidityTypePoolReserves)
	key = append(key, liquidityTypeBytes...)
	key = append(key, []byte("/")...)

	feeBytes := sdk.Uint64ToBigEndian(p.Fee)
	key = append(key, feeBytes...)
	key = append(key, []byte("/")...)

	return key
}

func (p PoolReservesKey) Counterpart() *PoolReservesKey {
	feeInt64 := utils.MustSafeUint64ToInt64(p.Fee)
	return &PoolReservesKey{
		TradePairId:           p.TradePairId.Reversed(),
		TickIndexTakerToMaker: p.TickIndexTakerToMaker*-1 + 2*feeInt64,
		Fee:                   p.Fee,
	}
}

func (p PoolReservesKey) Price() (priceTakerToMaker math_utils.PrecDec, err error) {
	return CalcPrice(p.TickIndexTakerToMaker)
}

func (p PoolReservesKey) MustPrice() (priceTakerToMaker math_utils.PrecDec) {
	price, err := p.Price()
	if err != nil {
		panic(err)
	}
	return price
}

func (p PoolReservesKey) PriceTakerToMaker() (priceTakerToMaker math_utils.PrecDec, err error) {
	return CalcPrice(-p.TickIndexTakerToMaker)
}
