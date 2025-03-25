package types

import (
	"cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/v6/app/params"
)

func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdktypes.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}
	if err := msg.Params.Validate(); err != nil {
		return errors.Wrap(err, "params are invalid")
	}
	if msg.Params.RewardAsset != params.DefaultDenom {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "reward asset change is prohibited")
	}
	return nil
}

func (msg *MsgFundTreasury) Validate() error {
	if len(msg.Amount) != 1 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "exactly one coin must be provided")
	}
	if msg.Amount[0].Denom != RewardDenom {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "provided denom doesn't match the reward denom %s", RewardDenom)
	}
	if err := msg.Amount.Validate(); err != nil {
		return errors.Wrap(err, "invalid coins")
	}
	return nil
}
