package types

import (
	"fmt"

	"cosmossdk.io/errors"

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

	if err := msg.checkHooksUnique(); err != nil {
		return fmt.Errorf("subscription hooks are not unique: %v", err)
	}

	if err := msg.checkHooksExist(); err != nil {
		return fmt.Errorf("non-existing hook type: %v", err)
	}

	return nil
}

func (msg *MsgManageHookSubscription) checkHooksUnique() error {
	cache := make(map[string]bool)
	for _, item := range msg.HookSubscription.Hooks {
		if cache[item.String()] {
			return fmt.Errorf("non-unique hook=%s", item.String())
		}
		cache[item.String()] = true
	}

	return nil
}

func (msg *MsgManageHookSubscription) checkHooksExist() error {
	for _, item := range msg.HookSubscription.Hooks {
		if err := ValidateHookType(item); err != nil {
			return err
		}
	}
	return nil
}
