package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateSchedule = "create_schedule"
	TypeMsgUpdateSchedule = "update_schedule"
	TypeMsgDeleteSchedule = "delete_schedule"
)

var _ sdk.Msg = &MsgCreateSchedule{}

func NewMsgCreateSchedule(
    creator string,
    index string,
    name string,
    period string,
    msgs string,
    
) *MsgCreateSchedule {
  return &MsgCreateSchedule{
		Creator : creator,
		Index: index,
		Name: name,
        Period: period,
        Msgs: msgs,
        
	}
}

func (msg *MsgCreateSchedule) Route() string {
  return RouterKey
}

func (msg *MsgCreateSchedule) Type() string {
  return TypeMsgCreateSchedule
}

func (msg *MsgCreateSchedule) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgCreateSchedule) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateSchedule) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  	if err != nil {
  		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  	}
  return nil
}

var _ sdk.Msg = &MsgUpdateSchedule{}

func NewMsgUpdateSchedule(
    creator string,
    index string,
    name string,
    period string,
    msgs string,
    
) *MsgUpdateSchedule {
  return &MsgUpdateSchedule{
		Creator: creator,
        Index: index,
        Name: name,
        Period: period,
        Msgs: msgs,
        
	}
}

func (msg *MsgUpdateSchedule) Route() string {
  return RouterKey
}

func (msg *MsgUpdateSchedule) Type() string {
  return TypeMsgUpdateSchedule
}

func (msg *MsgUpdateSchedule) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateSchedule) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateSchedule) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  }
   return nil
}

var _ sdk.Msg = &MsgDeleteSchedule{}

func NewMsgDeleteSchedule(
    creator string,
    index string,
    
) *MsgDeleteSchedule {
  return &MsgDeleteSchedule{
		Creator: creator,
		Index: index,
        
	}
}
func (msg *MsgDeleteSchedule) Route() string {
  return RouterKey
}

func (msg *MsgDeleteSchedule) Type() string {
  return TypeMsgDeleteSchedule
}

func (msg *MsgDeleteSchedule) GetSigners() []sdk.AccAddress {
  creator, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteSchedule) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteSchedule) ValidateBasic() error {
  _, err := sdk.AccAddressFromBech32(msg.Creator)
  if err != nil {
    return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
  }
  return nil
}