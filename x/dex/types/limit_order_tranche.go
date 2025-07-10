package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/utils"
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
		ReservesMakerDenom:    reservesMakerDenom,
		ReservesTakerDenom:    reservesTakerDenom,
		DecReservesMakerDenom: math_utils.NewPrecDecFromInt(reservesMakerDenom),
		DecReservesTakerDenom: math_utils.NewPrecDecFromInt(reservesTakerDenom),
		TotalMakerDenom:       totalMakerDenom,
		TotalTakerDenom:       totalTakerDenom,
		DecTotalTakerDenom:    math_utils.NewPrecDecFromInt(totalTakerDenom),
		MakerPrice:            makerPrice,
		PriceTakerToMaker:     math_utils.OnePrecDec().Quo(makerPrice),
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
	return t.DecReservesMakerDenom.Equal(math_utils.NewPrecDecFromInt(t.TotalMakerDenom))
}

func (t LimitOrderTranche) IsFilled() bool {
	return t.DecReservesMakerDenom.IsZero()
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
	return t.DecReservesMakerDenom.IsPositive()
}

func (t LimitOrderTranche) HasTokenOut() bool {
	return t.DecReservesTakerDenom.IsPositive()
}

func (t LimitOrderTranche) Price() math_utils.PrecDec {
	return t.MakerPrice
}

func (t LimitOrderTranche) RatioFilled() math_utils.PrecDec {
	amountFilled := t.DecTotalTakerDenom.Quo(t.MakerPrice)
	ratioFilled := amountFilled.QuoInt(t.TotalMakerDenom)

	// Cap ratio filled at 100% so that makers cannot over withdraw
	return math_utils.MinPrecDec(ratioFilled, math_utils.OnePrecDec())
}

func (t LimitOrderTranche) AmountUnfilled() math_utils.PrecDec {
	amountFilled := t.DecTotalTakerDenom.Quo(t.MakerPrice)
	trueAmountUnfilled := math_utils.NewPrecDecFromInt(t.TotalMakerDenom).Sub(amountFilled)

	// It is possible for a tranche to be overfilled due to rounding. Thus we cap the unfilled amount at 0
	withdrawableAmount := math_utils.MaxPrecDec(math_utils.ZeroPrecDec(), trueAmountUnfilled)
	return withdrawableAmount
}

func (t *LimitOrderTranche) RemoveTokenIn(
	trancheUser *LimitOrderTrancheUser,
) (amountToRemove math_utils.PrecDec) {
	amountUnfilled := t.AmountUnfilled()
	amountToRemove = amountUnfilled.MulInt(trancheUser.SharesOwned).QuoInt(t.TotalMakerDenom)
	t.SetMakerReserves(t.DecReservesMakerDenom.Sub(amountToRemove))

	return amountToRemove
}

func (t *LimitOrderTranche) CalcWithdrawAmount(trancheUser *LimitOrderTrancheUser) (sharesToWithdraw math.Int, tokenOut math_utils.PrecDec) {
	ratioFilled := t.RatioFilled()
	maxAllowedToWithdraw := ratioFilled.MulInt(trancheUser.SharesOwned)
	sharesToWithdrawDec := maxAllowedToWithdraw.Sub(math_utils.NewPrecDecFromInt(trancheUser.SharesWithdrawn))

	// Given rounding it is possible for sharesToWithdrawn > maxAllowedToWithdraw. In this case we just exit.
	if !sharesToWithdrawDec.IsPositive() {
		return math.ZeroInt(), math_utils.ZeroPrecDec()
	}
	amountOutTokenOutDec := sharesToWithdrawDec.Mul(t.MakerPrice)

	// Round shares withdrawn up to ensure math favors dex
	return sharesToWithdrawDec.Ceil().TruncateInt(), amountOutTokenOutDec
}

func (t *LimitOrderTranche) Withdraw(trancheUser *LimitOrderTrancheUser) (sharesWithdrawn math.Int, tokenOut math_utils.PrecDec) {
	amountOutTokenIn, amountOutTokenOut := t.CalcWithdrawAmount(trancheUser)
	t.SetTakerReserves(t.DecReservesTakerDenom.Sub(amountOutTokenOut))

	return amountOutTokenIn, amountOutTokenOut
}

func (t *LimitOrderTranche) Swap(maxAmountTakerIn math_utils.PrecDec, maxAmountMakerOut *math_utils.PrecDec) (
	inAmount math_utils.PrecDec,
	outAmount math_utils.PrecDec,
) {
	reservesTokenOut := t.DecReservesMakerDenom
	fillTokenIn := t.DecReservesTakerDenom

	maxOutGivenIn := maxAmountTakerIn.Quo(t.MakerPrice)
	possibleOutAmounts := []math_utils.PrecDec{reservesTokenOut, maxOutGivenIn}
	if maxAmountMakerOut != nil {
		possibleOutAmounts = append(possibleOutAmounts, *maxAmountMakerOut)
	}
	outAmount = utils.MinPrecDecArr(possibleOutAmounts)

	// Due to precision loss when when doing division before multipliation the amountIn can be greater than maxAmountTakerIn
	// so we need to cap it at maxAmountTakerIn
	inAmount = math_utils.MinPrecDec(
		t.MakerPrice.Mul(outAmount),
		maxAmountTakerIn,
	)

	t.SetTakerReserves(fillTokenIn.Add(inAmount))
	t.SetMakerReserves(reservesTokenOut.Sub(outAmount))
	t.SetTotalTakerDenom(t.DecTotalTakerDenom.Add(inAmount))

	return inAmount, outAmount
}

func (t *LimitOrderTranche) PlaceMakerLimitOrder(amountIn math.Int) {
	t.SetMakerReserves(t.DecReservesMakerDenom.Add(math_utils.NewPrecDecFromInt(amountIn)))
	t.TotalMakerDenom = t.TotalMakerDenom.Add(amountIn)
}

func (t *LimitOrderTranche) SetMakerReserves(reserves math_utils.PrecDec) {
	t.ReservesMakerDenom = reserves.TruncateInt()
	t.DecReservesMakerDenom = reserves
}

func (t *LimitOrderTranche) SetTakerReserves(reserves math_utils.PrecDec) {
	t.ReservesTakerDenom = reserves.TruncateInt()
	t.DecReservesTakerDenom = reserves
}

func (t *LimitOrderTranche) SetTotalTakerDenom(reserves math_utils.PrecDec) {
	t.TotalTakerDenom = reserves.TruncateInt()
	t.DecTotalTakerDenom = reserves
}
