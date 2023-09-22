package types

import (
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/utils/math"
	"github.com/neutron-org/neutron/x/dex/utils"
)

var _ TickLiquidityKey = (*PoolReservesKey)(nil)

func (p PoolReservesKey) KeyMarshal() []byte {
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

	liquidityTypeBytes := []byte(LiquidityTypePoolReserves)
	key = append(key, liquidityTypeBytes...)
	key = append(key, []byte("/")...)

	feeBytes := sdk.Uint64ToBigEndian(p.Fee)
	key = append(key, feeBytes...)
	key = append(key, []byte("/")...)

	return key
}

func (p PoolReservesKey) KeyUnmarshal(bz []byte) error {
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

	if split[3] != LiquidityTypePoolReserves {
		return errors.New("unexpected liquidity type")
	}

	p.Fee = sdk.BigEndianToUint64([]byte(split[4]))

	return nil
}

func (p PoolReservesKey) Counterpart() *PoolReservesKey {
	feeInt64 := utils.MustSafeUint64ToInt64(p.Fee)
	return &PoolReservesKey{
		TradePairID:           p.TradePairID.Reversed(),
		TickIndexTakerToMaker: p.TickIndexTakerToMaker*-1 + 2*feeInt64,
		Fee:                   p.Fee,
	}
}

func (p PoolReservesKey) PriceTakerToMaker() (priceTakerToMaker math_utils.PrecDec, err error) {
	return CalcPrice(p.TickIndexTakerToMaker)
}

func (p PoolReservesKey) MustPriceTakerToMaker() (priceTakerToMaker math_utils.PrecDec) {
	price, err := p.PriceTakerToMaker()
	if err != nil {
		panic(err)
	}
	return price
}
