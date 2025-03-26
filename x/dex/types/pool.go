package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
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

	maxOutGivenTakerIn := math_utils.NewPrecDecFromInt(maxAmountTakerIn).Quo(makerReserves.MakerPrice).TruncateInt()
	possibleAmountsMakerOut := []math.Int{makerReserves.ReservesMakerDenom, maxOutGivenTakerIn}
	if maxAmountMakerOut != nil {
		possibleAmountsMakerOut = append(possibleAmountsMakerOut, *maxAmountMakerOut)
	}

	// outAmount will be the smallest value of:
	// a) The available reserves1
	// b) The most the user could get out given maxAmountIn0 (maxOutGivenIn1)
	// c) The maximum amount the user wants out (maxAmountOut1)
	amountMakerOut = utils.MinIntArr(possibleAmountsMakerOut)

	amountTakerIn = makerReserves.MakerPrice.MulInt(amountMakerOut).Ceil().TruncateInt()
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

	centerPrice1To0 := p.MustCalcPrice1To0Center()
	var depositValueAsToken0 math_utils.PrecDec
	autoswapFee := math_utils.ZeroPrecDec()

	if !autoswap {
		inAmount0, inAmount1 = CalcGreatestMatchingRatio(
			*lowerReserve0,
			*upperReserve1,
			maxAmount0,
			maxAmount1,
		)
		depositValueAsToken0 = CalcAmountAsToken0(inAmount0, inAmount1, centerPrice1To0)

	} else {
		residualAmount0, residualAmount1 := CalcAutoswapAmount(
			*lowerReserve0, *upperReserve1,
			maxAmount0, maxAmount1,
			centerPrice1To0,
		)

		residualDepositValueAsToken0 := CalcAmountAsToken0(residualAmount0, residualAmount1, centerPrice1To0)
		autoswapFee = p.CalcAutoswapFee(residualDepositValueAsToken0)

		fullDepositValueAsToken0 := CalcAmountAsToken0(maxAmount0, maxAmount1, centerPrice1To0)
		depositValueAsToken0 = fullDepositValueAsToken0.Sub(autoswapFee)

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
		return p.UpperTick1.MakerPrice
	}

	return p.LowerTick0.MakerPrice
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

// Balance deposit amounts to match the existing ratio in the pool. If pool is empty allow any ratio.
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

// CalcAutoswapAmount calculates the smallest swap to match the current pool ratio.
// see: https://www.notion.so/Autoswap-Spec-ca5f35a4cd5b4dbf9ae27e0454ddd445?pvs=4#12032ea59b0e802c925efae10c3ca85f
func CalcAutoswapAmount(
	reserves0,
	reserves1,
	depositAmount0,
	depositAmount1 math.Int,
	price1To0 math_utils.PrecDec,
) (resultAmount0, resultAmount1 math.Int) {
	if reserves0.IsZero() && reserves1.IsZero() {
		// The pool is empty, any deposit amount is allowed. Nothing to be swapped
		return math.ZeroInt(), math.ZeroInt()
	}

	reserves0Dec := math_utils.NewPrecDecFromInt(reserves0)
	reserves1Dec := math_utils.NewPrecDecFromInt(reserves1)
	// swapAmount = (reserves0*depositAmount1 - reserves1*depositAmount0) / (price * reserves1  + reserves0)
	swapAmount := reserves0Dec.MulInt(depositAmount1).Sub(reserves1Dec.MulInt(depositAmount0)).
		Quo(reserves0Dec.Add(reserves1Dec.Quo(price1To0)))

	switch {
	case swapAmount.IsZero(): // nothing to be swapped
		return math.ZeroInt(), math.ZeroInt()

	case swapAmount.IsPositive(): // Token1 needs to be swapped
		return math.ZeroInt(), swapAmount.Ceil().TruncateInt()

	default: // Token0 needs to be swapped
		amountSwappedAs1 := swapAmount.Neg()

		amountSwapped0 := amountSwappedAs1.Quo(price1To0)
		return amountSwapped0.Ceil().TruncateInt(), math.ZeroInt()
	}
}

func (p *Pool) CalcAutoswapFee(depositValueAsToken0 math_utils.PrecDec) math_utils.PrecDec {
	feeInt64 := utils.MustSafeUint64ToInt64(p.Fee())
	feeAsPrice := MustCalcPrice(-feeInt64)
	autoSwapFee := math_utils.OnePrecDec().Sub(feeAsPrice)

	// fee = depositValueAsToken0 * (1 - p(fee) )
	return autoSwapFee.Mul(depositValueAsToken0)
}

func CalcAmountAsToken0(amount0, amount1 math.Int, price1To0 math_utils.PrecDec) math_utils.PrecDec {
	amount0Dec := math_utils.NewPrecDecFromInt(amount0)
	amount1Dec := math_utils.NewPrecDecFromInt(amount1)

	return amount0Dec.Add(amount1Dec.Quo(price1To0))
}
