package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v2/utils/math"
	"github.com/neutron-org/neutron/v2/x/dex/utils"
)

func NewPool(
	pairID *PairID,
	centerTickIndexNormalized int64,
	fee uint64,
	id uint64,
) (*Pool, error) {
	feeInt64 := utils.MustSafeUint64ToInt64(fee)

	id0To1 := &PoolReservesKey{
		TradePairId:           NewTradePairIDFromMaker(pairID, pairID.Token1),
		TickIndexTakerToMaker: centerTickIndexNormalized + feeInt64,
		Fee:                   fee,
	}

	upperTick, err := NewPoolReserves(id0To1)
	if err != nil {
		return nil, err
	}

	lowerTick := NewPoolReservesFromCounterpart(upperTick)

	return &Pool{
		LowerTick0: lowerTick,
		UpperTick1: upperTick,
		Id:         id,
	}, nil
}

func MustNewPool(
	pairID *PairID,
	centerTickIndexNormalized int64,
	fee uint64,
	id uint64,
) *Pool {
	pool, err := NewPool(pairID, centerTickIndexNormalized, fee, id)
	if err != nil {
		panic("Error while creating new pool: " + err.Error())
	}
	return pool
}

func (p *Pool) CenterTickIndex() int64 {
	feeInt64 := utils.MustSafeUint64ToInt64(p.Fee())
	return p.UpperTick1.Key.TickIndexTakerToMaker - feeInt64
}

func (p *Pool) Fee() uint64 {
	return p.UpperTick1.Key.Fee
}

func (p *Pool) GetLowerReserve0() math.Int {
	return p.LowerTick0.ReservesMakerDenom
}

func (p *Pool) GetUpperReserve1() math.Int {
	return p.UpperTick1.ReservesMakerDenom
}

func (p *Pool) Swap(
	tradePairID *TradePairID,
	maxAmountTakerIn math.Int,
	maxAmountMakerOut *math.Int,
) (amountTakerIn, amountMakerOut math.Int) {
	var takerReserves, makerReserves *PoolReserves
	if tradePairID.IsMakerDenomToken0() {
		makerReserves = p.LowerTick0
		takerReserves = p.UpperTick1
	} else {
		makerReserves = p.UpperTick1
		takerReserves = p.LowerTick0
	}

	if maxAmountTakerIn.Equal(math.ZeroInt()) ||
		makerReserves.ReservesMakerDenom.Equal(math.ZeroInt()) {
		return math.ZeroInt(), math.ZeroInt()
	}

	maxOutGivenTakerIn := makerReserves.PriceTakerToMaker.MulInt(maxAmountTakerIn).TruncateInt()
	possibleAmountsMakerOut := []math.Int{makerReserves.ReservesMakerDenom, maxOutGivenTakerIn}
	if maxAmountMakerOut != nil {
		possibleAmountsMakerOut = append(possibleAmountsMakerOut, *maxAmountMakerOut)
	}

	// outAmount will be the smallest value of:
	// a.) The available reserves1
	// b.) The most the user could get out given maxAmountIn0 (maxOutGivenIn1)
	// c.) The maximum amount the user wants out (maxAmountOut1)
	amountMakerOut = utils.MinIntArr(possibleAmountsMakerOut)
	amountTakerIn = math_utils.NewPrecDecFromInt(amountMakerOut).
		Quo(makerReserves.PriceTakerToMaker).
		TruncateInt()

	takerReserves.ReservesMakerDenom = takerReserves.ReservesMakerDenom.Add(amountTakerIn)
	makerReserves.ReservesMakerDenom = makerReserves.ReservesMakerDenom.Sub(amountMakerOut)

	return amountTakerIn, amountMakerOut
}

// Mutates the Pool object and returns relevant change variables. Deposit is not committed until
// pool.save() is called or the underlying ticks are saved; this method does not use any keeper methods.
func (p *Pool) Deposit(
	maxAmount0,
	maxAmount1,
	existingShares math.Int,
	autoswap bool,
) (inAmount0, inAmount1 math.Int, outShares sdk.Coin) {
	lowerReserve0 := &p.LowerTick0.ReservesMakerDenom
	upperReserve1 := &p.UpperTick1.ReservesMakerDenom

	inAmount0, inAmount1 = CalcGreatestMatchingRatio(
		*lowerReserve0,
		*upperReserve1,
		maxAmount0,
		maxAmount1,
	)

	if inAmount0.Equal(math.ZeroInt()) && inAmount1.Equal(math.ZeroInt()) {
		return math.ZeroInt(), math.ZeroInt(), sdk.Coin{Denom: p.GetPoolDenom()}
	}

	outShares = p.CalcSharesMinted(inAmount0, inAmount1, existingShares)

	if autoswap {
		residualAmount0 := maxAmount0.Sub(inAmount0)
		residualAmount1 := maxAmount1.Sub(inAmount1)

		// NOTE: Currently not doing anything with the error,
		// but added error handling to all of the new functions for autoswap.
		// Open to changing it however.
		residualShares, _ := p.CalcResidualSharesMinted(residualAmount0, residualAmount1)

		outShares = outShares.Add(residualShares)

		inAmount0 = maxAmount0
		inAmount1 = maxAmount1
	}

	*lowerReserve0 = lowerReserve0.Add(inAmount0)
	*upperReserve1 = upperReserve1.Add(inAmount1)

	return inAmount0, inAmount1, outShares
}

func (p *Pool) GetPoolDenom() string {
	return NewPoolDenom(p.Id)
}

func (p *Pool) Price(tradePairID *TradePairID) math_utils.PrecDec {
	if tradePairID.IsTakerDenomToken0() {
		return p.UpperTick1.PriceTakerToMaker
	}

	return p.LowerTick0.PriceTakerToMaker
}

func (p *Pool) MustCalcPrice1To0Center() math_utils.PrecDec {
	// NOTE: We can safely call the error-less version of CalcPrice here because the pool object
	// has already been initialized with an upper and lower tick which satisfy a check for IsTickOutOfRange
	return MustCalcPrice(-1 * p.CenterTickIndex())
}

func (p *Pool) CalcSharesMinted(
	amount0 math.Int,
	amount1 math.Int,
	existingShares math.Int,
) (sharesMinted sdk.Coin) {
	price1To0Center := p.MustCalcPrice1To0Center()
	valueMintedToken0 := CalcAmountAsToken0(amount0, amount1, price1To0Center)

	valueExistingToken0 := CalcAmountAsToken0(
		p.LowerTick0.ReservesMakerDenom,
		p.UpperTick1.ReservesMakerDenom,
		price1To0Center,
	)
	var sharesMintedAmount math.Int
	if valueExistingToken0.GT(math_utils.ZeroPrecDec()) {
		sharesMintedAmount = valueMintedToken0.MulInt(existingShares).
			Quo(valueExistingToken0).
			TruncateInt()
	} else {
		sharesMintedAmount = valueMintedToken0.TruncateInt()
	}

	return sdk.Coin{Denom: p.GetPoolDenom(), Amount: sharesMintedAmount}
}

func (p *Pool) CalcResidualSharesMinted(
	residualAmount0 math.Int,
	residualAmount1 math.Int,
) (sharesMinted sdk.Coin, err error) {
	fee := CalcFee(p.UpperTick1.Key.TickIndexTakerToMaker, p.LowerTick0.Key.TickIndexTakerToMaker)
	valueMintedToken0, err := CalcResidualValue(
		residualAmount0,
		residualAmount1,
		p.LowerTick0.PriceTakerToMaker,
		fee,
	)
	if err != nil {
		return sdk.Coin{Denom: p.GetPoolDenom()}, err
	}

	return sdk.Coin{Denom: p.GetPoolDenom(), Amount: valueMintedToken0.TruncateInt()}, nil
}

func (p *Pool) RedeemValue(sharesToRemove, totalShares math.Int) (outAmount0, outAmount1 math.Int) {
	reserves0 := &p.LowerTick0.ReservesMakerDenom
	reserves1 := &p.UpperTick1.ReservesMakerDenom
	// outAmount1 = ownershipRatio * reserves1
	//            = (sharesToRemove / totalShares) * reserves1
	//            = (reserves1 * sharesToRemove ) / totalShares
	outAmount1 = math.LegacyNewDecFromInt(reserves1.Mul(sharesToRemove)).QuoInt(totalShares).TruncateInt()
	// outAmount0 = ownershipRatio * reserves1
	//            = (sharesToRemove / totalShares) * reserves1
	//            = (reserves1 * sharesToRemove ) / totalShares
	outAmount0 = math.LegacyNewDecFromInt(reserves0.Mul(sharesToRemove)).QuoInt(totalShares).TruncateInt()

	return outAmount0, outAmount1
}

func (p *Pool) Withdraw(sharesToRemove, totalShares math.Int) (outAmount0, outAmount1 math.Int) {
	reserves0 := &p.LowerTick0.ReservesMakerDenom
	reserves1 := &p.UpperTick1.ReservesMakerDenom
	outAmount0, outAmount1 = p.RedeemValue(sharesToRemove, totalShares)
	*reserves0 = reserves0.Sub(outAmount0)
	*reserves1 = reserves1.Sub(outAmount1)

	return outAmount0, outAmount1
}

// Balance trueAmount1 to the pool ratio
func CalcGreatestMatchingRatio(
	targetAmount0 math.Int,
	targetAmount1 math.Int,
	amount0 math.Int,
	amount1 math.Int,
) (resultAmount0, resultAmount1 math.Int) {
	targetAmount0Dec := math.LegacyNewDecFromInt(targetAmount0)
	targetAmount1Dec := math.LegacyNewDecFromInt(targetAmount1)

	// See spec: https://www.notion.so/dualityxyz/Autoswap-Spec-e856fa7b2438403c95147010d479b98c
	if targetAmount1.GT(math.ZeroInt()) {
		resultAmount0 = math.MinInt(
			amount0,
			math.LegacyNewDecFromInt(amount1).Mul(targetAmount0Dec).Quo(targetAmount1Dec).TruncateInt())
	} else {
		resultAmount0 = amount0
	}

	if targetAmount0.GT(math.ZeroInt()) {
		resultAmount1 = math.MinInt(
			amount1,
			math.LegacyNewDecFromInt(amount0).Mul(targetAmount1Dec).Quo(targetAmount0Dec).TruncateInt())
	} else {
		resultAmount1 = amount1
	}

	return resultAmount0, resultAmount1
}

func CalcResidualValue(
	amount0, amount1 math.Int,
	priceLowerTakerToMaker math_utils.PrecDec,
	fee int64,
) (math_utils.PrecDec, error) {
	// ResidualValue = Amount0 * (Price1to0Center / Price1to0Upper) + Amount1 * Price1to0Lower
	amount0Discount, err := CalcPrice(-fee)
	if err != nil {
		return math_utils.ZeroPrecDec(), err
	}

	return amount0Discount.MulInt(amount0).Add(priceLowerTakerToMaker.MulInt(amount1)), nil
}

func CalcFee(upperTickIndex, lowerTickIndex int64) int64 {
	return (upperTickIndex - lowerTickIndex) / 2
}

func CalcAmountAsToken0(amount0, amount1 math.Int, price1To0 math_utils.PrecDec) math_utils.PrecDec {
	amount0Dec := math_utils.NewPrecDecFromInt(amount0)

	return amount0Dec.Add(price1To0.MulInt(amount1))
}
