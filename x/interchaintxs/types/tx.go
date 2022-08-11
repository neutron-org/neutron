package types

import (
	"errors"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_ codectypes.UnpackInterfacesMessage = MsgSubmitTx{}
)

func (m *MsgRegisterInterchainAccount) ValidateBasic() error {
	if len(m.ConnectionId) == 0 {
		return errors.New("empty connection id")
	}

	if _, err := sdk.AccAddressFromBech32(m.FromAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse FromAddress: %s", m.FromAddress)
	}

	if len(m.InterchainAccountId) == 0 {
		return errors.New("empty interchainAccountID")
	}

	return nil
}

func (m *MsgRegisterInterchainAccount) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(m.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

func (m *MsgRegisterInterchainAccount) Route() string {
	return RouterKey
}

func (m *MsgRegisterInterchainAccount) Type() string {
	return "register-interchain-account"
}

func (m MsgRegisterInterchainAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

//----------------------------------------------------------------

func (m *MsgSubmitTx) ValidateBasic() error {
	if len(m.ConnectionId) == 0 {
		return errors.New("empty connection id")
	}

	if _, err := sdk.AccAddressFromBech32(m.FromAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse FromAddress: %s", m.FromAddress)
	}

	if len(m.InterchainAccountId) == 0 {
		return errors.New("empty interchainAccountID")
	}

	if len(m.Msgs) == 0 {
		return errors.New("no messages provided")
	}

	return nil
}

func (m *MsgSubmitTx) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(m.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

func (m *MsgSubmitTx) Route() string {
	return RouterKey
}

func (m *MsgSubmitTx) Type() string {
	return "submit-tx"
}

// GetTxMsgs casts the attached *types.Any messages to SDK messages.
func (m *MsgSubmitTx) GetTxMsgs() (sdkMsgs []sdk.Msg, err error) {
	for idx, msg := range m.Msgs {
		sdkMsg, ok := msg.GetCachedValue().(sdk.Msg)
		if !ok {
			return nil, fmt.Errorf("failed to cast message #%d to sdk.Msg", idx)
		}

		sdkMsgs = append(sdkMsgs, sdkMsg)
	}

	return
}

func (m MsgSubmitTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
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
	var (
		sdkMsg sdk.Msg
	)
	for _, m := range msg.Msgs {
		if err := unpacker.UnpackAny(m, &sdkMsg); err != nil {
			return err
		}
	}
	return nil
}
