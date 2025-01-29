package types

import (
	"cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdktypes.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}
	if err := msg.Params.Validate(); err != nil {
		return errors.Wrap(err, "params are invalid")
	}
	return nil
}

func (msg *MsgFundTreasury) Validate(params Params) error {
	if len(msg.Amount) != 1 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "exactly one coin must be provided")
	}
	if msg.Amount[0].Denom != params.DenomCompensation {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "provided denom doesn't match the denom for compensation %s", params.DenomCompensation)
	}
	if err := msg.Amount.Validate(); err != nil {
		return errors.Wrap(err, "invalid coins")
	}
	return nil
}
