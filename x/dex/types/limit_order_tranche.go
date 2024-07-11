package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/utils"
)

func NewLimitOrderTranche(
	makerDenom string,
	takerDenom string,
	trancheKey string,
	tickIndex int64,
	reservesMakerDenom math.Int,
	reservesTakerDenom math.Int,
	totalMakerDenom math.Int,
	totalTakerDenom math.Int,
) (*LimitOrderTranche, error) {
	tradePairID, err := NewTradePairID(takerDenom, makerDenom)
	if err != nil {
		return nil, err
	}
	makerPrice, err := tradePairID.MakerPrice(tickIndex)
	if err != nil {
		return nil, err
	}
	return &LimitOrderTranche{
		Key: &LimitOrderTrancheKey{
			TradePairId:           tradePairID,
			TrancheKey:            trancheKey,
			TickIndexTakerToMaker: tickIndex,
		},
		ReservesMakerDenom: reservesMakerDenom,
		ReservesTakerDenom: reservesTakerDenom,
		TotalMakerDenom:    totalMakerDenom,
		TotalTakerDenom:    totalTakerDenom,
		MakerPrice:         makerPrice,
	}, nil
}

// Useful for testing
func MustNewLimitOrderTranche(
	makerDenom string,
	takerDenom string,
	trancheKey string,
	tickIndex int64,
	reservesMakerDenom math.Int,
	reservesTakerDenom math.Int,
	totalMakerDenom math.Int,
	totalTakerDenom math.Int,
	expirationTime ...time.Time,
) *LimitOrderTranche {
	limitOrderTranche, err := NewLimitOrderTranche(
		makerDenom,
		takerDenom,
		trancheKey,
		tickIndex,
		reservesMakerDenom,
		reservesTakerDenom,
		totalMakerDenom,
		totalTakerDenom,
	)
	if err != nil {
		panic(err)
	}
	switch len(expirationTime) {
	case 0:
		break
	case 1:
		limitOrderTranche.ExpirationTime = &expirationTime[0]
	default:
		panic("can only supply one expiration time")
	}
	return limitOrderTranche
}

func (t LimitOrderTranche) IsPlaceTranche() bool {
	return t.ReservesMakerDenom.Equal(t.TotalMakerDenom)
}

func (t LimitOrderTranche) IsFilled() bool {
	return t.ReservesMakerDenom.IsZero()
}

func (t LimitOrderTranche) HasExpiration() bool {
	return t.ExpirationTime != nil
}

func (t LimitOrderTranche) IsJIT() bool {
	return t.ExpirationTime != nil && *t.ExpirationTime == JITGoodTilTime()
}

func (t LimitOrderTranche) IsExpired(ctx sdk.Context) bool {
	return t.ExpirationTime != nil && !t.IsJIT() && !t.ExpirationTime.After(ctx.BlockTime())
}

func (t LimitOrderTranche) HasTokenIn() bool {
	return t.ReservesMakerDenom.GT(math.ZeroInt())
}

func (t LimitOrderTranche) HasTokenOut() bool {
	return t.ReservesTakerDenom.GT(math.ZeroInt())
}

func (t LimitOrderTranche) Price() math_utils.PrecDec {
	return t.MakerPrice
}

func (t LimitOrderTranche) RatioFilled() math_utils.PrecDec {
	amountFilled := t.PriceTakerToMaker.MulInt(t.TotalTakerDenom)
	ratioFilled := amountFilled.QuoInt(t.TotalMakerDenom)

	// Cap ratio filled at 100% so that makers cannot over withdraw
	return math_utils.MinPrecDec(ratioFilled, math_utils.OnePrecDec())
}

func (t LimitOrderTranche) AmountUnfilled() math_utils.PrecDec {
	amountFilled := t.PriceTakerToMaker.MulInt(t.TotalTakerDenom)
	trueAmountUnfilled := math_utils.NewPrecDecFromInt(t.TotalMakerDenom).Sub(amountFilled)

	// It is possible for a tranche to be overfilled due to rounding. Thus we cap the unfilled amount at 0
	withdrawableAmount := math_utils.MaxPrecDec(math_utils.ZeroPrecDec(), trueAmountUnfilled)
	return withdrawableAmount
}

func (t LimitOrderTranche) HasLiquidity() bool {
	return t.ReservesMakerDenom.GT(math.ZeroInt())
}

func (t *LimitOrderTranche) RemoveTokenIn(
	trancheUser *LimitOrderTrancheUser,
) (amountToRemove math.Int) {
	amountUnfilled := t.AmountUnfilled()
	maxAmountToRemove := amountUnfilled.MulInt(trancheUser.SharesOwned).
		QuoInt(t.TotalMakerDenom).
		TruncateInt()
	amountToRemove = maxAmountToRemove.Sub(trancheUser.SharesCancelled)
	t.ReservesMakerDenom = t.ReservesMakerDenom.Sub(amountToRemove)

	return amountToRemove
}

func (t *LimitOrderTranche) CalcWithdrawAmount(trancheUser *LimitOrderTrancheUser) (sharesToWithdraw, tokenOut math.Int) {
	ratioFilled := t.RatioFilled()
	maxAllowedToWithdraw := ratioFilled.MulInt(trancheUser.SharesOwned)
	sharesToWithdrawDec := maxAllowedToWithdraw.Sub(math_utils.NewPrecDecFromInt(trancheUser.SharesWithdrawn))
	amountOutTokenOutDec := sharesToWithdrawDec.Quo(t.PriceTakerToMaker)

	// Round shares withdrawn up and amountOut down to ensure math favors dex
	return sharesToWithdrawDec.Ceil().TruncateInt(), amountOutTokenOutDec.TruncateInt()
}

func (t *LimitOrderTranche) Withdraw(trancheUser *LimitOrderTrancheUser) (sharesWithdrawn, tokenOut math.Int) {
	amountOutTokenIn, amountOutTokenOut := t.CalcWithdrawAmount(trancheUser)
	t.ReservesTakerDenom = t.ReservesTakerDenom.Sub(amountOutTokenOut)

	return amountOutTokenIn, amountOutTokenOut
}

func (t *LimitOrderTranche) Swap(maxAmountTakerIn math.Int, maxAmountMakerOut *math.Int) (
	inAmount math.Int,
	outAmount math.Int,
) {
	reservesTokenOut := &t.ReservesMakerDenom
	fillTokenIn := &t.ReservesTakerDenom
	totalTokenIn := &t.TotalTakerDenom
	maxOutGivenIn := t.PriceTakerToMaker.MulInt(maxAmountTakerIn).TruncateInt()
	possibleOutAmounts := []math.Int{*reservesTokenOut, maxOutGivenIn}
	if maxAmountMakerOut != nil {
		possibleOutAmounts = append(possibleOutAmounts, *maxAmountMakerOut)
	}
	outAmount = utils.MinIntArr(possibleOutAmounts)

	inAmount = math_utils.NewPrecDecFromInt(outAmount).Quo(t.PriceTakerToMaker).Ceil().TruncateInt()

	*fillTokenIn = fillTokenIn.Add(inAmount)
	*totalTokenIn = totalTokenIn.Add(inAmount)
	*reservesTokenOut = reservesTokenOut.Sub(outAmount)

	return inAmount, outAmount
}

func (t *LimitOrderTranche) PlaceMakerLimitOrder(amountIn math.Int) {
	t.ReservesMakerDenom = t.ReservesMakerDenom.Add(amountIn)
	t.TotalMakerDenom = t.TotalMakerDenom.Add(amountIn)
}
