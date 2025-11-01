package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// constants
const (
	TypeMsgCreateDenom       = "create_denom"
	TypeMsgMint              = "tf_mint"
	TypeMsgBurn              = "tf_burn"
	TypeMsgForceTransfer     = "force_transfer"
	TypeMsgChangeAdmin       = "change_admin"
	TypeMsgSetDenomMetadata  = "set_denom_metadata"
	TypeMsgSetBeforeSendHook = "set_before_send_hook"
)

var _ sdk.Msg = &MsgCreateDenom{}

// NewMsgCreateDenom creates a msg to create a new denom
func NewMsgCreateDenom(sender, subdenom string) *MsgCreateDenom {
	return &MsgCreateDenom{
		Sender:   sender,
		Subdenom: subdenom,
	}
}

func (m MsgCreateDenom) Route() string { return RouterKey }
func (m MsgCreateDenom) Type() string  { return TypeMsgCreateDenom }
func (m MsgCreateDenom) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = GetTokenDenom(m.Sender, m.Subdenom)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidDenom, err.Error())
	}

	return nil
}

func (m MsgCreateDenom) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgCreateDenom) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgMint{}

// NewMsgMint creates a message to mint tokens
func NewMsgMint(sender string, amount sdk.Coin) *MsgMint {
	return &MsgMint{
		Sender: sender,
		Amount: amount,
	}
}

func NewMsgMintTo(sender string, amount sdk.Coin, mintToAddress string) *MsgMint {
	return &MsgMint{
		Sender:        sender,
		Amount:        amount,
		MintToAddress: mintToAddress,
	}
}

func (m MsgMint) Route() string { return RouterKey }
func (m MsgMint) Type() string  { return TypeMsgMint }
func (m MsgMint) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if m.MintToAddress != "" {
		if _, err := sdk.AccAddressFromBech32(m.MintToAddress); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid mint_to_address (%s)", err)
		}
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(math.ZeroInt()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgMint) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgMint) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgBurn{}

// NewMsgBurn creates a message to burn tokens
func NewMsgBurn(sender string, amount sdk.Coin) *MsgBurn {
	return &MsgBurn{
		Sender: sender,
		Amount: amount,
	}
}

// NewMsgBurn creates a message to burn tokens
func NewMsgBurnFrom(sender string, amount sdk.Coin, from string) *MsgBurn {
	return &MsgBurn{
		Sender:          sender,
		Amount:          amount,
		BurnFromAddress: from,
	}
}

func (m MsgBurn) Route() string { return RouterKey }
func (m MsgBurn) Type() string  { return TypeMsgBurn }
func (m MsgBurn) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if m.BurnFromAddress != "" {
		if _, err := sdk.AccAddressFromBech32(m.BurnFromAddress); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid burn_from_address (%s)", err)
		}
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(math.ZeroInt()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgBurn) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgBurn) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgForceTransfer{}

// NewMsgForceTransfer creates a transfer funds from one account to another
func NewMsgForceTransfer(sender string, amount sdk.Coin, fromAddr, toAddr string) *MsgForceTransfer {
	return &MsgForceTransfer{
		Sender:              sender,
		Amount:              amount,
		TransferFromAddress: fromAddr,
		TransferToAddress:   toAddr,
	}
}

func (m MsgForceTransfer) Route() string { return RouterKey }
func (m MsgForceTransfer) Type() string  { return TypeMsgForceTransfer }
func (m MsgForceTransfer) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.TransferFromAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(m.TransferToAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid address (%s)", err)
	}

	if !m.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

func (m MsgForceTransfer) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgForceTransfer) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgChangeAdmin{}

// NewMsgChangeAdmin creates a message to burn tokens
func NewMsgChangeAdmin(sender, denom, newAdmin string) *MsgChangeAdmin {
	return &MsgChangeAdmin{
		Sender:   sender,
		Denom:    denom,
		NewAdmin: newAdmin,
	}
}

func (m MsgChangeAdmin) Route() string { return RouterKey }
func (m MsgChangeAdmin) Type() string  { return TypeMsgChangeAdmin }
func (m MsgChangeAdmin) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid address (%s)", err)
	}

	_, _, err = DeconstructDenom(m.Denom)
	if err != nil {
		return err
	}

	return nil
}

func (m MsgChangeAdmin) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgChangeAdmin) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSetDenomMetadata{}

// NewMsgSetDenomMetadata creates a message to set a metadata for a denom
func NewMsgSetDenomMetadata(sender string, metadata banktypes.Metadata) *MsgSetDenomMetadata {
	return &MsgSetDenomMetadata{
		Sender:   sender,
		Metadata: metadata,
	}
}

func (m MsgSetDenomMetadata) Route() string { return RouterKey }
func (m MsgSetDenomMetadata) Type() string  { return TypeMsgSetDenomMetadata }
func (m MsgSetDenomMetadata) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = m.Metadata.Validate()
	if err != nil {
		return err
	}

	_, _, err = DeconstructDenom(m.Metadata.Base)
	if err != nil {
		return err
	}

	return nil
}

func (m MsgSetDenomMetadata) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgSetDenomMetadata) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSetBeforeSendHook{}

// NewMsgSetBeforeSendHook creates a message to set a new before send hook
func NewMsgSetBeforeSendHook(sender, denom, contractAddr string) *MsgSetBeforeSendHook {
	return &MsgSetBeforeSendHook{
		Sender:       sender,
		Denom:        denom,
		ContractAddr: contractAddr,
	}
}

func (m MsgSetBeforeSendHook) Route() string { return RouterKey }
func (m MsgSetBeforeSendHook) Type() string  { return TypeMsgSetBeforeSendHook }
func (m MsgSetBeforeSendHook) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if m.ContractAddr != "" {
		_, err = sdk.AccAddressFromBech32(m.ContractAddr)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid cosmwasm contract address (%s)", err)
		}
	}

	_, _, err = DeconstructDenom(m.Denom)
	if err != nil {
		return ErrInvalidDenom
	}

	return nil
}

func (m MsgSetBeforeSendHook) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&m)
}

func (m MsgSetBeforeSendHook) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}
