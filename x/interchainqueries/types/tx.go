package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
)

var (
	_ codectypes.UnpackInterfacesMessage = MsgSubmitQueryResult{}
)

func (msg MsgSubmitQueryResult) Route() string {
	return RouterKey
}

func (msg MsgSubmitQueryResult) Type() string {
	return "submit-query-result"
}

func (msg MsgSubmitQueryResult) ValidateBasic() error {
	// TODO: add basic validation.
	return nil
}

func (msg MsgSubmitQueryResult) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))

}

func (msg MsgSubmitQueryResult) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgRegisterInterchainQuery) Route() string {
	return RouterKey
}

func (msg MsgRegisterInterchainQuery) Type() string {
	return "register-interchain-query"
}

func (msg MsgRegisterInterchainQuery) ValidateBasic() error {
	// TODO
	return nil
}

func (msg MsgRegisterInterchainQuery) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))

}

func (msg MsgRegisterInterchainQuery) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgSubmitQueryResult) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, b := range msg.Result.Blocks {
		var header exported.Header
		if err := unpacker.UnpackAny(b.Header, &header); err != nil {
			return err
		}
		if err := unpacker.UnpackAny(b.NextBlockHeader, &header); err != nil {
			return err
		}
	}
	return nil
}
