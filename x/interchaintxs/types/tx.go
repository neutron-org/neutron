package types

import (
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

func (m *MsgRegisterInterchainAccount) ValidateBasic() error {
	if len(m.ConnectionId) == 0 {
		return errors.New("empty connection id")
	}

	if len(m.Owner) == 0 {
		return errors.New("empty owner")
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

	if len(m.Owner) == 0 {
		return errors.New("empty owner")
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
