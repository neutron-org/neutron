package types

import (
	"cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//----------------------------------------------------------------

var _ sdk.Msg = &MsgManageHookSubscription{}

func (msg *MsgManageHookSubscription) Route() string {
	return RouterKey
}

func (msg *MsgManageHookSubscription) Type() string {
	return "manage-hook-subscription"
}

func (msg *MsgManageHookSubscription) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgManageHookSubscription) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(msg)
}

func (msg *MsgManageHookSubscription) Validate() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}

	if _, err := sdk.AccAddressFromBech32(msg.HookSubscription.ContractAddress); err != nil {
		return errors.Wrap(err, "hook subscription contractAddress is invalid")
	}

	if !msg.areHooksUnique() {
		return fmt.Errorf("subscription hooks are not unique")
	}

	if !msg.allHooksExist() {
		return fmt.Errorf("non-existing hook type")
	}

	return nil
}

func (msg *MsgManageHookSubscription) areHooksUnique() bool {
	cache := make(map[string]bool)
	for _, item := range msg.HookSubscription.Hooks {
		if cache[item.String()] {
			return false
		}
		cache[item.String()] = true
	}

	return true
}

func (msg *MsgManageHookSubscription) allHooksExist() bool {
	for _, item := range msg.HookSubscription.Hooks {
		_, ok := HookType_name[int32(item)]
		if !ok {
			return false
		}
	}

	return true
}

//----------------------------------------------------------------

var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateParams) Type() string {
	return "update-params"
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(msg)
}

func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}

	return nil
}
