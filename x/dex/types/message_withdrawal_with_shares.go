package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgWithdrawalWithShares = "withdrawal_with_shares"

var _ sdk.Msg = &MsgWithdrawalWithShares{}

func NewMsgWithdrawalWithShares(
	creator,
	receiver string,
	sharesToRemove sdk.Coins,
) *MsgWithdrawalWithShares {
	return &MsgWithdrawalWithShares{
		Creator:        creator,
		Receiver:       receiver,
		SharesToRemove: sharesToRemove,
	}
}

func (msg *MsgWithdrawalWithShares) Route() string {
	return RouterKey
}

func (msg *MsgWithdrawalWithShares) Type() string {
	return TypeMsgWithdrawalWithShares
}

func (msg *MsgWithdrawalWithShares) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgWithdrawalWithShares) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return bz
}

func (msg *MsgWithdrawalWithShares) Validate() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	if len(msg.SharesToRemove) == 0 {
		return sdkerrors.Wrapf(ErrZeroWithdraw, "shares to remove is empty")
	}

	poolsSeen := make(map[string]bool)
	for _, share := range msg.SharesToRemove {

		if !share.Amount.IsPositive() {
			return sdkerrors.Wrapf(ErrZeroWithdraw, "share amount is zero for share %s", share.Denom)
		}

		if err := ValidatePoolDenom(share.Denom); err != nil {
			return sdkerrors.Wrapf(ErrInvalidPoolDenom, "invalid share denom (%s)", err)
		}
		if _, ok := poolsSeen[share.Denom]; ok {
			return sdkerrors.Wrapf(ErrDuplicatePoolWithdraw, "pool with denom %s is duplicated", share.Denom)
		}
		poolsSeen[share.Denom] = true
	}

	return nil
}
