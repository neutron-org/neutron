package types

import (
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Shared Attributes
const (
	AttributeCreator              = "Creator"
	AttributeReceiver             = "Receiver"
	AttributeToken0               = "TokenZero"
	AttributeToken1               = "TokenOne"
	AttributeTokenIn              = "TokenIn"
	AttributeTokenOut             = "TokenOut"
	AttributeAmountIn             = "AmountIn"
	AttributeAmountOut            = "AmountOut"
	AttributeSwapAmountIn         = "SwapAmountIn"
	AttributeSwapAmountOut        = "SwapAmountOut"
	AttributeTokenInAmountOut     = "TokenInAmountOut"
	AttributeTokenOutAmountOut    = "TokenOutAmountOut"
	AttributeTickIndex            = "TickIndex"
	AttributeFee                  = "Fee"
	AttributeTrancheKey           = "TrancheKey"
	AttributeSharesMinted         = "SharesMinted"
	AttributeReserves0Deposited   = "ReservesZeroDeposited"
	AttributeReserves1Deposited   = "ReservesOneDeposited"
	AttributeReserves0Withdrawn   = "ReservesZeroWithdrawn"
	AttributeReserves1Withdrawn   = "ReservesOneWithdrawn"
	AttributeSharesRemoved        = "SharesRemoved"
	AttributeRoute                = "Route"
	AttributeDust                 = "Dust"
	AttributeLimitTick            = "LimitTick"
	AttributeOrderType            = "OrderType"
	AttributeShares               = "Shares"
	AttributeReserves             = "Reserves"
	AttributeGas                  = "Gas"
	AttributeDenom                = "denom"
	AttributeWithdrawn            = "total_withdrawn"
	AttributeGasConsumed          = "gas_consumed"
	AttributeLiquidityTickType    = "liquidity_tick_type"
	AttributeLp                   = "lp"
	AttributeLimitOrder           = "limit_order"
	AttributeIsExpiringLimitOrder = "is_expiring_limit_order"
	AttributeInc                  = "inc"
	AttributeDec                  = "dec"
	AttributePairID               = "pair_id"
	AttributeTimestamp            = "Timestamp"
)

// Event Keys
const (
	DepositEventKey                  = "DepositLP"
	WithdrawEventKey                 = "WithdrawLP"
	MultihopSwapEventKey             = "MultihopSwap"
	PlaceLimitOrderEventKey          = "PlaceLimitOrder"
	WithdrawFilledLimitOrderEventKey = "WithdrawLimitOrder"
	CancelLimitOrderEventKey         = "CancelLimitOrder"
	EventTypeTickUpdate              = "TickUpdate"
	TickUpdateEventKey               = "TickUpdate"
	EventTypeGoodTilPurgeHitGasLimit = "GoodTilPurgeHitGasLimit"
	// EventTypeNeutronMessage defines the event type used by the Interchain Queries module events.
	EventTypeNeutronMessage = "neutron"
)

func EmitEventWithTimestamp(ctx sdk.Context, event sdk.Event) {
	timestamp := sdk.NewAttribute(AttributeTimestamp, ctx.BlockTime().Format(time.RFC3339))
	event = event.AppendAttributes(timestamp)
	ctx.EventManager().EmitEvent(event)
}

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
		sdk.NewAttribute(AttributeCreator, creator.String()),
		sdk.NewAttribute(AttributeReceiver, receiver.String()),
		sdk.NewAttribute(AttributeToken0, token0),
		sdk.NewAttribute(AttributeToken1, token1),
		sdk.NewAttribute(AttributeTickIndex, strconv.FormatInt(tickIndex, 10)),
		sdk.NewAttribute(AttributeFee, strconv.FormatUint(fee, 10)),
		sdk.NewAttribute(AttributeReserves0Deposited, depositAmountReserve0.String()),
		sdk.NewAttribute(AttributeReserves1Deposited, depositAmountReserve1.String()),
		sdk.NewAttribute(AttributeSharesMinted, sharesMinted.String()),
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
		sdk.NewAttribute(AttributeCreator, creator.String()),
		sdk.NewAttribute(AttributeReceiver, receiver.String()),
		sdk.NewAttribute(AttributeToken0, token0),
		sdk.NewAttribute(AttributeToken1, token1),
		sdk.NewAttribute(AttributeTickIndex, strconv.FormatInt(tickIndex, 10)),
		sdk.NewAttribute(AttributeFee, strconv.FormatUint(fee, 10)),
		sdk.NewAttribute(AttributeReserves0Withdrawn, withdrawAmountReserve0.String()),
		sdk.NewAttribute(AttributeReserves1Withdrawn, withdrawAmountReserve1.String()),
		sdk.NewAttribute(AttributeSharesRemoved, sharesRemoved.String()),
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
		sdk.NewAttribute(AttributeCreator, creator.String()),
		sdk.NewAttribute(AttributeReceiver, receiver.String()),
		sdk.NewAttribute(AttributeTokenIn, makerDenom),
		sdk.NewAttribute(AttributeTokenOut, tokenOut),
		sdk.NewAttribute(AttributeAmountIn, amountIn.String()),
		sdk.NewAttribute(AttributeAmountOut, amountOut.String()),
		sdk.NewAttribute(AttributeRoute, strings.Join(route, ",")),
		sdk.NewAttribute(AttributeDust, strings.Join(dustStrings, ",")),
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
	swapAmountIn math.Int,
	swapAmountOut math.Int,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, PlaceLimitOrderEventKey),
		sdk.NewAttribute(AttributeCreator, creator.String()),
		sdk.NewAttribute(AttributeReceiver, receiver.String()),
		sdk.NewAttribute(AttributeToken0, token0),
		sdk.NewAttribute(AttributeToken1, token1),
		sdk.NewAttribute(AttributeTokenIn, makerDenom),
		sdk.NewAttribute(AttributeTokenOut, tokenOut),
		sdk.NewAttribute(AttributeAmountIn, amountIn.String()),
		sdk.NewAttribute(AttributeLimitTick, strconv.FormatInt(limitTick, 10)),
		sdk.NewAttribute(AttributeOrderType, orderType),
		sdk.NewAttribute(AttributeShares, shares.String()),
		sdk.NewAttribute(AttributeTrancheKey, trancheKey),
		sdk.NewAttribute(AttributeSwapAmountIn, swapAmountIn.String()),
		sdk.NewAttribute(AttributeSwapAmountOut, swapAmountOut.String()),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func WithdrawFilledLimitOrderEvent(
	creator sdk.AccAddress,
	token0 string,
	token1 string,
	makerDenom string,
	tokenOut string,
	amountOutTaker math.Int,
	amountOutMaker math.Int,
	trancheKey string,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, WithdrawFilledLimitOrderEventKey),
		sdk.NewAttribute(AttributeCreator, creator.String()),
		sdk.NewAttribute(AttributeReceiver, creator.String()),
		sdk.NewAttribute(AttributeToken0, token0),
		sdk.NewAttribute(AttributeToken1, token1),
		sdk.NewAttribute(AttributeTokenIn, makerDenom),
		sdk.NewAttribute(AttributeTokenOut, tokenOut),
		sdk.NewAttribute(AttributeTrancheKey, trancheKey),
		// DEPRECATED: `AmountOut` will be removed in the next release
		sdk.NewAttribute(AttributeAmountOut, amountOutTaker.String()),
		sdk.NewAttribute(AttributeTokenOutAmountOut, amountOutTaker.String()),
		sdk.NewAttribute(AttributeTokenInAmountOut, amountOutMaker.String()),
	}

	return sdk.NewEvent(sdk.EventTypeMessage, attrs...)
}

func CancelLimitOrderEvent(
	creator sdk.AccAddress,
	token0 string,
	token1 string,
	makerDenom string,
	tokenOut string,
	amountOutTaker math.Int,
	amountOutMaker math.Int,
	trancheKey string,
) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(sdk.AttributeKeyAction, CancelLimitOrderEventKey),
		sdk.NewAttribute(AttributeCreator, creator.String()),
		sdk.NewAttribute(AttributeReceiver, creator.String()),
		sdk.NewAttribute(AttributeToken0, token0),
		sdk.NewAttribute(AttributeToken1, token1),
		sdk.NewAttribute(AttributeTokenIn, makerDenom),
		sdk.NewAttribute(AttributeTokenOut, tokenOut),
		sdk.NewAttribute(AttributeTokenOutAmountOut, amountOutTaker.String()),
		sdk.NewAttribute(AttributeTokenInAmountOut, amountOutMaker.String()),
		sdk.NewAttribute(AttributeTrancheKey, trancheKey),
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
		sdk.NewAttribute(AttributeToken0, token0),
		sdk.NewAttribute(AttributeToken1, token1),
		sdk.NewAttribute(AttributeTokenIn, makerDenom),
		sdk.NewAttribute(AttributeTickIndex, strconv.FormatInt(tickIndex, 10)),
		sdk.NewAttribute(AttributeReserves, reserves.String()),
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
		sdk.NewAttribute(AttributeFee, strconv.FormatUint(tick.Key.Fee, 10)),
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
		sdk.NewAttribute(AttributeTrancheKey, tranche.Key.TrancheKey),
	)
}

func CreateTickUpdateLimitOrderTranchePurge(tranche *LimitOrderTranche) sdk.Event {
	tradePairID := tranche.Key.TradePairId
	pairID := tradePairID.MustPairID()
	return TickUpdateEvent(
		pairID.Token0,
		pairID.Token1,
		tradePairID.MakerDenom,
		tranche.Key.TickIndexTakerToMaker,
		math.ZeroInt(),
		sdk.NewAttribute(AttributeTrancheKey, tranche.Key.TrancheKey),
	)
}

func GoodTilPurgeHitLimitEvent(gas types.Gas) sdk.Event {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, "dex"),
		sdk.NewAttribute(AttributeGas, strconv.FormatUint(gas, 10)),
	}

	return sdk.NewEvent(EventTypeGoodTilPurgeHitGasLimit, attrs...)
}

func GetEventsWithdrawnAmount(coins sdk.Coins) sdk.Events {
	events := sdk.Events{}
	for _, coin := range coins {
		event := sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(AttributeDenom, coin.Denom),
			sdk.NewAttribute(AttributeWithdrawn, coin.Amount.String()),
		)
		events = append(events, event)
	}
	return events
}

func GetEventsGasConsumed(gasBefore, gasAfter types.Gas) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(AttributeGasConsumed, strconv.FormatUint(gasAfter-gasBefore, 10)),
		),
	}
}

func GetEventsIncExpiringOrders(pairID *TradePairID) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeInc),
			sdk.NewAttribute(AttributeLiquidityTickType, AttributeLimitOrder),
			sdk.NewAttribute(AttributeIsExpiringLimitOrder, strconv.FormatBool(true)),
			sdk.NewAttribute(AttributePairID, pairID.String()),
		),
	}
}

func GetEventsDecExpiringOrders(pairID *TradePairID) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeDec),
			sdk.NewAttribute(AttributeLiquidityTickType, AttributeLimitOrder),
			sdk.NewAttribute(AttributeIsExpiringLimitOrder, strconv.FormatBool(true)),
			sdk.NewAttribute(AttributePairID, pairID.String()),
		),
	}
}

func GetEventsIncTotalOrders(pairID *TradePairID) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeInc),
			sdk.NewAttribute(AttributeLiquidityTickType, AttributeLimitOrder),
			sdk.NewAttribute(AttributeIsExpiringLimitOrder, strconv.FormatBool(false)),
			sdk.NewAttribute(AttributePairID, pairID.String()),
		),
	}
}

func GetEventsDecTotalOrders(pairID *TradePairID) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeDec),
			sdk.NewAttribute(AttributeLiquidityTickType, AttributeLimitOrder),
			sdk.NewAttribute(AttributeIsExpiringLimitOrder, strconv.FormatBool(false)),
			sdk.NewAttribute(AttributePairID, pairID.String()),
		),
	}
}

func GetEventsIncTotalPoolReserves(pairID PairID) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeInc),
			sdk.NewAttribute(AttributeLiquidityTickType, AttributeLp),
			sdk.NewAttribute(AttributeIsExpiringLimitOrder, strconv.FormatBool(false)),
			sdk.NewAttribute(AttributePairID, pairID.String()),
		),
	}
}

func GetEventsDecTotalPoolReserves(pairID PairID) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeDec),
			sdk.NewAttribute(AttributeLiquidityTickType, AttributeLp),
			sdk.NewAttribute(AttributeIsExpiringLimitOrder, strconv.FormatBool(false)),
			sdk.NewAttribute(AttributePairID, pairID.String()),
		),
	}
}
