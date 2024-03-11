package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v3/utils/math"
	"github.com/neutron-org/neutron/v3/x/dex/utils"
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
	priceTakerToMaker, err := tradePairID.PriceTakerToMaker(tickIndex)
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
		PriceTakerToMaker:  priceTakerToMaker,
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
	return limitOrderTranche
}

func (t LimitOrderTranche) IsPlaceTranche() bool {
	return t.ReservesMakerDenom.Equal(t.TotalMakerDenom)
}

func (t LimitOrderTranche) IsFilled() bool {
	return t.ReservesMakerDenom.IsZero()
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
	return t.PriceTakerToMaker
}

func (t LimitOrderTranche) RatioFilled() math_utils.PrecDec {
	amountFilled := t.PriceTakerToMaker.MulInt(t.TotalTakerDenom)
	ratioFilled := amountFilled.QuoInt(t.TotalMakerDenom)
	return ratioFilled
}

func (t LimitOrderTranche) AmountUnfilled() math_utils.PrecDec {
	amountFilled := t.PriceTakerToMaker.MulInt(t.TotalTakerDenom)
	return math_utils.NewPrecDecFromInt(t.TotalMakerDenom).Sub(amountFilled)
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

func (t *LimitOrderTranche) Withdraw(trancheUser *LimitOrderTrancheUser) (math.Int, math_utils.PrecDec) {
	reservesTokenOutDec := math_utils.NewPrecDecFromInt(t.ReservesTakerDenom)

	ratioFilled := t.RatioFilled()
	maxAllowedToWithdraw := ratioFilled.MulInt(trancheUser.SharesOwned).TruncateInt()
	amountOutTokenIn := maxAllowedToWithdraw.Sub(trancheUser.SharesWithdrawn)
	amountOutTokenOut := math_utils.NewPrecDecFromInt(amountOutTokenIn).Quo(t.PriceTakerToMaker)
	t.ReservesTakerDenom = reservesTokenOutDec.Sub(amountOutTokenOut).TruncateInt()

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

	inAmount = math_utils.NewPrecDecFromInt(outAmount).Quo(t.PriceTakerToMaker).TruncateInt()

	*fillTokenIn = fillTokenIn.Add(inAmount)
	*totalTokenIn = totalTokenIn.Add(inAmount)
	*reservesTokenOut = reservesTokenOut.Sub(outAmount)

	return inAmount, outAmount
}

func (t *LimitOrderTranche) PlaceMakerLimitOrder(amountIn math.Int) {
	t.ReservesMakerDenom = t.ReservesMakerDenom.Add(amountIn)
	t.TotalMakerDenom = t.TotalMakerDenom.Add(amountIn)
}
