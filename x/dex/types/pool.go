package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/utils"
)

type PoolShareholder struct {
	Address string
	Shares  math.Int
}

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

func (p *Pool) CenterTickIndexToken1() int64 {
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
	// a) The available reserves1
	// b) The most the user could get out given maxAmountIn0 (maxOutGivenIn1)
	// c) The maximum amount the user wants out (maxAmountOut1)
	amountMakerOut = utils.MinIntArr(possibleAmountsMakerOut)

	amountTakerIn = math_utils.NewPrecDecFromInt(amountMakerOut).
		Quo(makerReserves.PriceTakerToMaker).
		Ceil().
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

	centerPrice1To0 := p.MustCalcPrice1To0Center()
	depositValueAsToken0 := CalcAmountAsToken0(inAmount0, inAmount1, centerPrice1To0)
	autoswapFee := math_utils.ZeroPrecDec()
	if autoswap {
		residualAmount0 := maxAmount0.Sub(inAmount0)
		residualAmount1 := maxAmount1.Sub(inAmount1)

		// NOTE: Currently not doing anything with the error,
		// but added error handling to all of the new functions for autoswap.
		// Open to changing it however.
		residualDepositValueAsToken0 := CalcAmountAsToken0(residualAmount0, residualAmount1, centerPrice1To0)
		autoswapFee, _ = p.CalcAutoswapFee(residualDepositValueAsToken0)
		depositValueAsToken0 = depositValueAsToken0.Add(residualDepositValueAsToken0.Sub(autoswapFee))

		inAmount0 = maxAmount0
		inAmount1 = maxAmount1
	}

	outShares = p.CalcSharesMinted(depositValueAsToken0, existingShares, autoswapFee)
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
	return MustCalcPrice(-1 * p.CenterTickIndexToken1())
}

func (p *Pool) CalcSharesMinted(
	depositValueAsToken0 math_utils.PrecDec,
	existingShares math.Int,
	autoswapFee math_utils.PrecDec,
) (sharesMinted sdk.Coin) {
	price1To0Center := p.MustCalcPrice1To0Center()

	valueExistingToken0 := CalcAmountAsToken0(
		p.LowerTick0.ReservesMakerDenom,
		p.UpperTick1.ReservesMakerDenom,
		price1To0Center,
	)

	totalValueWithAutoswapFeeToken0 := valueExistingToken0.Add(autoswapFee)
	var sharesMintedAmount math.Int
	if valueExistingToken0.GT(math_utils.ZeroPrecDec()) {
		sharesMintedAmount = depositValueAsToken0.MulInt(existingShares).
			Quo(totalValueWithAutoswapFeeToken0).
			TruncateInt()
	} else {
		sharesMintedAmount = depositValueAsToken0.TruncateInt()
	}

	return sdk.Coin{Denom: p.GetPoolDenom(), Amount: sharesMintedAmount}
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

func (p *Pool) CalcAutoswapFee(
	depositValueAsToken0 math_utils.PrecDec,
) (math_utils.PrecDec, error) {
	feeInt64 := utils.MustSafeUint64ToInt64(p.Fee())
	feeAsPrice, err := CalcPrice(feeInt64)
	autoSwapFee := math_utils.OnePrecDec().Sub(feeAsPrice)
	if err != nil {
		return math_utils.ZeroPrecDec(), err
	}
	// fee = depositValueAsToken0 * (1 - p(fee) )
	return autoSwapFee.Mul(depositValueAsToken0), nil
}

func CalcAmountAsToken0(amount0, amount1 math.Int, price1To0 math_utils.PrecDec) math_utils.PrecDec {
	amount0Dec := math_utils.NewPrecDecFromInt(amount0)

	return amount0Dec.Add(price1To0.MulInt(amount1))
}
