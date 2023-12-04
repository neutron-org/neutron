package types

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	goodTil int64,
	maxAmountOut *math.Int,
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
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPlaceLimitOrder) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	if msg.AmountIn.LTE(math.ZeroInt()) {
		return ErrZeroLimitOrder
	}

	if msg.OrderType.IsGoodTil() && msg.ExpirationTime == 0 {
		return ErrGoodTilOrderWithoutExpiration
	}

	if !msg.OrderType.IsGoodTil() && msg.ExpirationTime != 0 {
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

	return nil
}

func (msg *MsgPlaceLimitOrder) ValidateGoodTilExpiration(curBlockTime int64) error {
	if msg.OrderType.IsGoodTil() && curBlockTime > msg.ExpirationTime {
		return sdkerrors.Wrapf(ErrExpirationTimeInPast,
			"Current BlockTime (%d) is after Provided ExpirationTime (%d)",
			curBlockTime,
			msg.ExpirationTime,
		)
	}

	return nil
}
