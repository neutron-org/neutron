package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/utils"
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

func (p *Pool) GetLowerReserve0() math_utils.PrecDec {
	return p.LowerTick0.DecReservesMakerDenom
}

func (p *Pool) GetUpperReserve1() math_utils.PrecDec {
	return p.UpperTick1.DecReservesMakerDenom
}

func (p *Pool) Swap(
	tradePairID *TradePairID,
	maxAmountTakerIn math_utils.PrecDec,
	maxAmountMakerOut *math_utils.PrecDec,
) (amountTakerIn, amountMakerOut math_utils.PrecDec) {
	var takerReserves, makerReserves *PoolReserves
	if tradePairID.IsMakerDenomToken0() {
		makerReserves = p.LowerTick0
		takerReserves = p.UpperTick1
	} else {
		makerReserves = p.UpperTick1
		takerReserves = p.LowerTick0
	}

	if maxAmountTakerIn.IsZero() ||
		makerReserves.DecReservesMakerDenom.IsZero() {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec()
	}

	maxOutGivenTakerIn := maxAmountTakerIn.Quo(makerReserves.MakerPrice)
	possibleAmountsMakerOut := []math_utils.PrecDec{makerReserves.DecReservesMakerDenom, maxOutGivenTakerIn}
	if maxAmountMakerOut != nil {
		possibleAmountsMakerOut = append(possibleAmountsMakerOut, *maxAmountMakerOut)
	}

	// outAmount will be the smallest value of:
	// a) The available reserves1
	// b) The most the user could get out given maxAmountIn0 (maxOutGivenIn1)
	// c) The maximum amount the user wants out (maxAmountOut1)
	amountMakerOut = utils.MinPrecDecArr(possibleAmountsMakerOut)

	// Due to precision loss when when doing division before multipliation the amountIn can be greater than maxAmountTakerIn
	// so we need to cap it at maxAmountTakerIn
	amountTakerIn = math_utils.MinPrecDec(
		makerReserves.MakerPrice.Mul(amountMakerOut),
		maxAmountTakerIn,
	)
	takerReserves.SetMakerReserves(takerReserves.DecReservesMakerDenom.Add(amountTakerIn))
	makerReserves.SetMakerReserves(makerReserves.DecReservesMakerDenom.Sub(amountMakerOut))

	return amountTakerIn, amountMakerOut
}

// Mutates the Pool object and returns relevant change variables. Deposit is not committed until
// pool.save() is called or the underlying ticks are saved; this method does not use any keeper methods.
func (p *Pool) Deposit(
	maxAmount0 math_utils.PrecDec,
	maxAmount1 math_utils.PrecDec,
	existingShares math.Int,
	autoswap bool,
) (inAmount0, inAmount1 math_utils.PrecDec, outShares sdk.Coin) {
	lowerReserve0 := &p.LowerTick0.DecReservesMakerDenom
	upperReserve1 := &p.UpperTick1.DecReservesMakerDenom

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
	p.LowerTick0.SetMakerReserves(lowerReserve0.Add(inAmount0))
	p.UpperTick1.SetMakerReserves(upperReserve1.Add(inAmount1))

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

	valueExistingToken0 := CalcDecAmountAsToken0(
		p.LowerTick0.DecReservesMakerDenom,
		p.UpperTick1.DecReservesMakerDenom,
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

func (p *Pool) RedeemValue(sharesToRemove, totalShares math.Int) (outAmount0, outAmount1 math_utils.PrecDec) {
	reserves0 := p.LowerTick0.DecReservesMakerDenom
	reserves1 := p.UpperTick1.DecReservesMakerDenom
	// outAmount1 = ownershipRatio * reserves1
	//            = (sharesToRemove / totalShares) * reserves1
	//            = (reserves1 * sharesToRemove ) / totalShares
	outAmount1 = math_utils.MinPrecDec(
		reserves1.MulInt(sharesToRemove).QuoInt(totalShares),
		reserves1,
	)
	// outAmount0 = ownershipRatio * reserves0
	//            = (sharesToRemove / totalShares) * reserves0
	//            = (reserves0 * sharesToRemove ) / totalShares
	outAmount0 = math_utils.MinPrecDec(
		reserves0.MulInt(sharesToRemove).QuoInt(totalShares),
		reserves0,
	)

	return outAmount0, outAmount1
}

func (p *Pool) Withdraw(sharesToRemove, totalShares math.Int) (outAmount0, outAmount1 math_utils.PrecDec) {
	outAmount0, outAmount1 = p.RedeemValue(sharesToRemove, totalShares)
	p.LowerTick0.SetMakerReserves(p.LowerTick0.DecReservesMakerDenom.Sub(outAmount0))
	p.UpperTick1.SetMakerReserves(p.UpperTick1.DecReservesMakerDenom.Sub(outAmount1))

	return outAmount0, outAmount1
}

// Balance deposit amounts to match the existing ratio in the pool. If pool is empty allow any ratio.
func CalcGreatestMatchingRatio(
	targetAmount0 math_utils.PrecDec,
	targetAmount1 math_utils.PrecDec,
	amount0 math_utils.PrecDec,
	amount1 math_utils.PrecDec,
) (resultAmount0, resultAmount1 math_utils.PrecDec) {
	if targetAmount1.IsPositive() {
		resultAmount0 = math_utils.MinPrecDec(
			amount0,
			amount1.Mul(targetAmount0).Quo(targetAmount1),
		)
	} else {
		resultAmount0 = amount0
	}

	if targetAmount0.IsPositive() {
		resultAmount1 = math_utils.MinPrecDec(
			amount1,
			amount0.Mul(targetAmount1).Quo(targetAmount0),
		)
	} else {
		resultAmount1 = amount1
	}

	return resultAmount0, resultAmount1
}

// CalcAutoswapAmount calculates the smallest swap to match the current pool ratio.
// see: https://www.notion.so/Autoswap-Spec-ca5f35a4cd5b4dbf9ae27e0454ddd445?pvs=4#12032ea59b0e802c925efae10c3ca85f
func CalcAutoswapAmount(
	reserves0 math_utils.PrecDec,
	reserves1 math_utils.PrecDec,
	depositAmount0 math_utils.PrecDec,
	depositAmount1 math_utils.PrecDec,
	price1To0 math_utils.PrecDec,
) (resultAmount0, resultAmount1 math_utils.PrecDec) {
	if reserves0.IsZero() && reserves1.IsZero() {
		// The pool is empty, any deposit amount is allowed. Nothing to be swapped
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec()
	}

	// swapAmount = (reserves0*depositAmount1 - reserves1*depositAmount0) / (price * reserves1  + reserves0)
	swapAmount := reserves0.Mul(depositAmount1).Sub(reserves1.Mul(depositAmount0)).
		Quo(reserves0.Add(reserves1.Quo(price1To0)))

	switch {
	case swapAmount.IsZero(): // nothing to be swapped
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec()

	case swapAmount.IsPositive(): // Token1 needs to be swapped
		return math_utils.ZeroPrecDec(), swapAmount.Ceil()

	default: // Token0 needs to be swapped
		amountSwappedAs1 := swapAmount.Neg()

		amountSwapped0 := amountSwappedAs1.Quo(price1To0)
		return amountSwapped0, math_utils.ZeroPrecDec()
	}
}

func (p *Pool) CalcAutoswapFee(depositValueAsToken0 math_utils.PrecDec) math_utils.PrecDec {
	feeInt64 := utils.MustSafeUint64ToInt64(p.Fee())
	feeAsPrice := MustCalcPrice(-feeInt64)
	autoSwapFee := math_utils.OnePrecDec().Sub(feeAsPrice)

	// fee = depositValueAsToken0 * (1 - p(fee) )
	return autoSwapFee.Mul(depositValueAsToken0)
}

func CalcAmountAsToken0(amount0, amount1, price1To0 math_utils.PrecDec) math_utils.PrecDec {
	return amount0.Add(amount1.Quo(price1To0))
}

func CalcDecAmountAsToken0(amount0, amount1, price1To0 math_utils.PrecDec) math_utils.PrecDec {
	return amount0.Add(amount1.Quo(price1To0))
}
