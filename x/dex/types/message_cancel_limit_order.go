package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgCancelLimitOrder = "cancel_limit_order"

var _ sdk.Msg = &MsgCancelLimitOrder{}

func NewMsgCancelLimitOrder(creator, trancheKey string) *MsgCancelLimitOrder {
	return &MsgCancelLimitOrder{
		Creator:    creator,
		TrancheKey: trancheKey,
	}
}

func (msg *MsgCancelLimitOrder) Route() string {
	return RouterKey
}

func (msg *MsgCancelLimitOrder) Type() string {
	return TypeMsgCancelLimitOrder
}

func (msg *MsgCancelLimitOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgCancelLimitOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return bz
}

func (msg *MsgCancelLimitOrder) Validate() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
