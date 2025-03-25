package types

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
)

const TypeMsgPlaceLimitOrder = "place_limit_order"

var _ sdk.Msg = &MsgPlaceLimitOrder{}

func NewMsgPlaceLimitOrder(
	creator,
	receiver,
	tokenIn,
	tokenOut string,
	tickIndex int64,
	amountIn math.Int,
	orderType LimitOrderType,
	goodTil *time.Time,
	maxAmountOut *math.Int,
	price *math_utils.PrecDec,
) *MsgPlaceLimitOrder {
	return &MsgPlaceLimitOrder{
		Creator:          creator,
		Receiver:         receiver,
		TokenIn:          tokenIn,
		TokenOut:         tokenOut,
		TickIndexInToOut: tickIndex,
		AmountIn:         amountIn,
		OrderType:        orderType,
		ExpirationTime:   goodTil,
		MaxAmountOut:     maxAmountOut,
		LimitSellPrice:   price,
	}
}

func (msg *MsgPlaceLimitOrder) Route() string {
	return RouterKey
}

func (msg *MsgPlaceLimitOrder) Type() string {
	return TypeMsgPlaceLimitOrder
}

func (msg *MsgPlaceLimitOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgPlaceLimitOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return bz
}

func (msg *MsgPlaceLimitOrder) Validate() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	// Verify tokenIn and tokenOut are valid denoms
	err = sdk.ValidateDenom(msg.TokenIn)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidDenom, "Error TokenIn denom (%s)", err)
	}

	err = sdk.ValidateDenom(msg.TokenOut)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidDenom, "Error TokenOut denom (%s)", err)
	}

	if msg.TokenIn == msg.TokenOut {
		return sdkerrors.Wrapf(ErrInvalidDenom, "tokenIn cannot equal tokenOut")
	}

	if msg.AmountIn.LTE(math.ZeroInt()) {
		return ErrZeroLimitOrder
	}

	if msg.OrderType.IsGoodTil() && msg.ExpirationTime == nil {
		return ErrGoodTilOrderWithoutExpiration
	}

	if !msg.OrderType.IsGoodTil() && msg.ExpirationTime != nil {
		return ErrExpirationOnWrongOrderType
	}

	if msg.MaxAmountOut != nil {
		if !msg.MaxAmountOut.IsPositive() {
			return ErrZeroMaxAmountOut
		}
		if !msg.OrderType.IsTakerOnly() {
			return ErrInvalidMaxAmountOutForMaker
		}
	}

	if IsTickOutOfRange(msg.TickIndexInToOut) {
		return ErrTickOutsideRange
	}

	if msg.LimitSellPrice != nil && IsPriceOutOfRange(*msg.LimitSellPrice) {
		return ErrPriceOutsideRange
	}

	if msg.LimitSellPrice != nil && msg.TickIndexInToOut != 0 {
		return ErrInvalidPriceAndTick
	}

	if msg.MinAverageSellPrice != nil && msg.MinAverageSellPrice.IsZero() {
		return ErrZeroMinAverageSellPrice
	}

	return nil
}

func (msg *MsgPlaceLimitOrder) ValidateGoodTilExpiration(blockTime time.Time) error {
	if msg.OrderType.IsGoodTil() && !msg.ExpirationTime.After(blockTime) {
		return sdkerrors.Wrapf(ErrExpirationTimeInPast,
			"Current BlockTime: %s; Provided ExpirationTime: %s",
			blockTime.String(),
			msg.ExpirationTime.String(),
		)
	}

	return nil
}
