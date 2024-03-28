package types

import (
	"strconv"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CreateDepositEvent(
	creator sdk.AccAddress,
	receiver sdk.AccAddress,
	token0 string,
	token1 string,
	tickIndex int64,
	fee uint64,
	depositAmountReserve0 math.Int,
	depositAmountReserve1 math.Int,
	sharesMinted math.Int,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, DepositEventKey),
		sdk.NewAttribute(DepositEventCreator, creator.String()),
		sdk.NewAttribute(DepositEventReceiver, receiver.String()),
		sdk.NewAttribute(DepositEventToken0, token0),
		sdk.NewAttribute(DepositEventToken1, token1),
		sdk.NewAttribute(DepositEventPrice, strconv.FormatInt(tickIndex, 10)),
		sdk.NewAttribute(DepositEventFee, strconv.FormatUint(fee, 10)),
		sdk.NewAttribute(DepositEventReserves0Deposited, depositAmountReserve0.String()),
		sdk.NewAttribute(DepositEventReserves1Deposited, depositAmountReserve1.String()),
		sdk.NewAttribute(DepositEventSharesMinted, sharesMinted.String()),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func CreateWithdrawEvent(
	creator sdk.AccAddress,
	receiver sdk.AccAddress,
	token0 string,
	token1 string,
	tickIndex int64,
	fee uint64,
	withdrawAmountReserve0 math.Int,
	withdrawAmountReserve1 math.Int,
	sharesRemoved math.Int,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, WithdrawEventKey),
		sdk.NewAttribute(WithdrawEventCreator, creator.String()),
		sdk.NewAttribute(WithdrawEventReceiver, receiver.String()),
		sdk.NewAttribute(WithdrawEventToken0, token0),
		sdk.NewAttribute(WithdrawEventToken1, token1),
		sdk.NewAttribute(WithdrawEventPrice, strconv.FormatInt(tickIndex, 10)),
		sdk.NewAttribute(WithdrawEventFee, strconv.FormatUint(fee, 10)),
		sdk.NewAttribute(WithdrawEventReserves0Withdrawn, withdrawAmountReserve0.String()),
		sdk.NewAttribute(WithdrawEventReserves1Withdrawn, withdrawAmountReserve1.String()),
		sdk.NewAttribute(WithdrawEventSharesRemoved, sharesRemoved.String()),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func CreateMultihopSwapEvent(
	creator sdk.AccAddress,
	receiver sdk.AccAddress,
	makerDenom string,
	tokenOut string,
	amountIn math.Int,
	amountOut math.Int,
	route []string,
	dust sdk.Coins,
) sdk.Event {
	dustStrings := make([]string, 0, dust.Len())
	for _, item := range dust {
		dustStrings = append(dustStrings, item.String())
	}
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, MultihopSwapEventKey),
		sdk.NewAttribute(MultihopSwapEventCreator, creator.String()),
		sdk.NewAttribute(MultihopSwapEventReceiver, receiver.String()),
		sdk.NewAttribute(MultihopSwapEventTokenIn, makerDenom),
		sdk.NewAttribute(MultihopSwapEventTokenOut, tokenOut),
		sdk.NewAttribute(MultihopSwapEventAmountIn, amountIn.String()),
		sdk.NewAttribute(MultihopSwapEventAmountOut, amountOut.String()),
		sdk.NewAttribute(MultihopSwapEventRoute, strings.Join(route, ",")),
		sdk.NewAttribute(MultihopSwapEventDust, strings.Join(dustStrings, ",")),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func CreatePlaceLimitOrderEvent(
	creator sdk.AccAddress,
	receiver sdk.AccAddress,
	token0 string,
	token1 string,
	makerDenom string,
	tokenOut string,
	amountIn math.Int,
	limitTick int64,
	orderType string,
	shares math.Int,
	trancheKey string,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, PlaceLimitOrderEventKey),
		sdk.NewAttribute(PlaceLimitOrderEventCreator, creator.String()),
		sdk.NewAttribute(PlaceLimitOrderEventReceiver, receiver.String()),
		sdk.NewAttribute(PlaceLimitOrderEventToken0, token0),
		sdk.NewAttribute(PlaceLimitOrderEventToken1, token1),
		sdk.NewAttribute(PlaceLimitOrderEventTokenIn, makerDenom),
		sdk.NewAttribute(PlaceLimitOrderEventTokenOut, tokenOut),
		sdk.NewAttribute(PlaceLimitOrderEventAmountIn, amountIn.String()),
		sdk.NewAttribute(PlaceLimitOrderEventLimitTick, strconv.FormatInt(limitTick, 10)),
		sdk.NewAttribute(PlaceLimitOrderEventOrderType, orderType),
		sdk.NewAttribute(PlaceLimitOrderEventShares, shares.String()),
		sdk.NewAttribute(PlaceLimitOrderEventTrancheKey, trancheKey),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func WithdrawFilledLimitOrderEvent(
	creator sdk.AccAddress,
	token0 string,
	token1 string,
	makerDenom string,
	tokenOut string,
	amountOut math.Int,
	trancheKey string,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, WithdrawFilledLimitOrderEventKey),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventCreator, creator.String()),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventToken0, token0),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventToken1, token1),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventTokenIn, makerDenom),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventTokenOut, tokenOut),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventTrancheKey, trancheKey),
		sdk.NewAttribute(WithdrawFilledLimitOrderEventAmountOut, amountOut.String()),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func CancelLimitOrderEvent(
	creator sdk.AccAddress,
	token0 string,
	token1 string,
	makerDenom string,
	tokenOut string,
	amountOut math.Int,
	trancheKey string,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, CancelLimitOrderEventKey),
		sdk.NewAttribute(CancelLimitOrderEventCreator, creator.String()),
		sdk.NewAttribute(CancelLimitOrderEventToken0, token0),
		sdk.NewAttribute(CancelLimitOrderEventToken1, token1),
		sdk.NewAttribute(CancelLimitOrderEventTokenIn, makerDenom),
		sdk.NewAttribute(CancelLimitOrderEventTokenOut, tokenOut),
		sdk.NewAttribute(CancelLimitOrderEventAmountOut, amountOut.String()),
		sdk.NewAttribute(CancelLimitOrderEventTrancheKey, trancheKey),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func TickUpdateEvent(
	token0 string,
	token1 string,
	makerDenom string,
	tickIndex int64,
	reserves math.Int,
	otherAttrs ...sdk.Attribute,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, TickUpdateEventKey),
		sdk.NewAttribute(TickUpdateEventToken0, token0),
		sdk.NewAttribute(TickUpdateEventToken1, token1),
		sdk.NewAttribute(TickUpdateEventTokenIn, makerDenom),
		sdk.NewAttribute(TickUpdateEventTickIndex, strconv.FormatInt(tickIndex, 10)),
		sdk.NewAttribute(TickUpdateEventFee, strconv.FormatInt(int64(0), 10)),
		sdk.NewAttribute(TickUpdateEventReserves, reserves.String()),
	}
	attrs = append(attrs, otherAttrs...)

	return sdk.NewEvent(EventTypeTickUpdate, attrs...)
}

func CreateTickUpdatePoolReserves(tick PoolReserves) sdk.Event {
	tradePairID := tick.Key.TradePairId
	pairID := tradePairID.MustPairID()
	return TickUpdateEvent(
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tick.Key.TickIndexTakerToMaker,
		tick.ReservesMakerDenom,
		sdk.NewAttribute(TickUpdateEventFee, strconv.FormatUint(tick.Key.Fee, 10)),
	)
}

func CreateTickUpdateLimitOrderTranche(tranche *LimitOrderTranche) sdk.Event {
	tradePairID := tranche.Key.TradePairId
	pairID := tradePairID.MustPairID()
	return TickUpdateEvent(
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tranche.Key.TickIndexTakerToMaker,
		tranche.ReservesMakerDenom,
		sdk.NewAttribute(TickUpdateEventTrancheKey, tranche.Key.TrancheKey),
	)
}

func GoodTilPurgeHitLimitEvent(gas sdk.Gas) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(GoodTilPurgeHitGasLimitEventGas, strconv.FormatUint(gas, 10)),
	}

	return sdk.NewEvent(EventTypeGoodTilPurgeHitGasLimit, attrs...)
}
