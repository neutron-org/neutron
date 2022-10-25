package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateFailure = "create_failure"
	TypeMsgUpdateFailure = "update_failure"
	TypeMsgDeleteFailure = "delete_failure"
)

var _ sdk.Msg = &MsgCreateFailure{}

func NewMsgCreateFailure(
    creator string,
    index string,
    contractAddress string,
    ackId string,
    ackType string,
    
) *MsgCreateFailure {
  return &MsgCreateFailure{
		Creator : creator,
		Index: index,
		ContractAddress: contractAddress,
        AckId: ackId,
        AckType: ackType,
        
	}
}

func (msg *MsgCreateFailure) Route() string {
  return RouterKey
}

func (msg *MsgCreateFailure) Type() string {
  return TypeMsgCreateFailure
}

func (msg *MsgCreateFailure) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgCreateFailure) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateFailure) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  	if err != nil {
  		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  	}
  return nil
}

var _ sdk.Msg = &MsgUpdateFailure{}

func NewMsgUpdateFailure(
    creator string,
    index string,
    contractAddress string,
    ackId string,
    ackType string,
    
) *MsgUpdateFailure {
  return &MsgUpdateFailure{
		Creator: creator,
        Index: index,
        ContractAddress: contractAddress,
        AckId: ackId,
        AckType: ackType,
        
	}
}

func (msg *MsgUpdateFailure) Route() string {
  return RouterKey
}

func (msg *MsgUpdateFailure) Type() string {
  return TypeMsgUpdateFailure
}

func (msg *MsgUpdateFailure) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateFailure) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateFailure) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  }
   return nil
}

var _ sdk.Msg = &MsgDeleteFailure{}

func NewMsgDeleteFailure(
    creator string,
    index string,
    
) *MsgDeleteFailure {
  return &MsgDeleteFailure{
		Creator: creator,
		Index: index,
        
	}
}
func (msg *MsgDeleteFailure) Route() string {
  return RouterKey
}

func (msg *MsgDeleteFailure) Type() string {
  return TypeMsgDeleteFailure
}

func (msg *MsgDeleteFailure) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteFailure) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteFailure) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  }
  return nil
}