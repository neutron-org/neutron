package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgWithdrawFilledLimitOrder = "withdrawal_withdrawn_limit_order"

var _ sdk.Msg = &MsgWithdrawFilledLimitOrder{}

func NewMsgWithdrawFilledLimitOrder(creator, trancheKey string) *MsgWithdrawFilledLimitOrder {
	return &MsgWithdrawFilledLimitOrder{
		Creator:    creator,
		TrancheKey: trancheKey,
	}
}

func (msg *MsgWithdrawFilledLimitOrder) Route() string {
	return RouterKey
}

func (msg *MsgWithdrawFilledLimitOrder) Type() string {
	return TypeMsgWithdrawFilledLimitOrder
}

func (msg *MsgWithdrawFilledLimitOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgWithdrawFilledLimitOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return bz
}

func (msg *MsgWithdrawFilledLimitOrder) Validate() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
