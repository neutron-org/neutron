package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

var _ codectypes.UnpackInterfacesMessage = MsgSubmitTx{}

func (msg *MsgRegisterInterchainAccount) ValidateBasic() error {
	if len(msg.ConnectionId) == 0 {
		return ErrEmptyConnectionID
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse FromAddress: %s", msg.FromAddress)
	}

	if len(msg.InterchainAccountId) == 0 {
		return ErrEmptyInterchainAccountID
	}

	return nil
}

func (msg *MsgRegisterInterchainAccount) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

func (msg *MsgRegisterInterchainAccount) Route() string {
	return RouterKey
}

func (msg *MsgRegisterInterchainAccount) Type() string {
	return "register-interchain-account"
}

func (msg MsgRegisterInterchainAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

//----------------------------------------------------------------

func (msg MsgSubmitTx) ValidateBasic() error {
	if err := msg.Fee.Validate(); err != nil {
		return err
	}

	if len(msg.ConnectionId) == 0 {
		return ErrEmptyConnectionID
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse FromAddress: %s", msg.FromAddress)
	}

	if len(msg.InterchainAccountId) == 0 {
		return ErrEmptyInterchainAccountID
	}

	if len(msg.Msgs) == 0 {
		return ErrNoMessages
	}

	if msg.Timeout <= 0 {
		return sdkerrors.Wrapf(ErrInvalidTimeout, "timeout must be greater than zero")
	}

	return nil
}

func (msg MsgSubmitTx) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

func (msg MsgSubmitTx) Route() string {
	return RouterKey
}

func (msg MsgSubmitTx) Type() string {
	return "submit-tx"
}

func (msg MsgSubmitTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// PackTxMsgAny marshals the sdk.Msg payload to a protobuf Any type
func PackTxMsgAny(sdkMsg sdk.Msg) (*codectypes.Any, error) {
	msg, ok := sdkMsg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", sdkMsg)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return any, nil
}

// implements UnpackInterfacesMessage.UnpackInterfaces (https://github.com/cosmos/cosmos-sdk/blob/d07d35f29e0a0824b489c552753e8798710ff5a8/codec/types/interface_registry.go#L60)
func (msg MsgSubmitTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var sdkMsg sdk.Msg
	for _, m := range msg.Msgs {
		if err := unpacker.UnpackAny(m, &sdkMsg); err != nil {
			return err
		}
	}
	return nil
}
