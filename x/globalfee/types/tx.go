package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority is invalid")
	}
	return nil
}
