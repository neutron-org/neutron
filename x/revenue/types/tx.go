package types

import (
	"cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	appparams "github.com/neutron-org/neutron/v6/app/params"
)

func (msg *MsgUpdateParams) Validate(params Params) error {
	if _, err := sdktypes.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}
	if err := msg.Params.Validate(); err != nil {
		return errors.Wrap(err, "params are invalid")
	}
	if msg.Params.RewardAsset != appparams.DefaultDenom {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "reward asset change is prohibited")
	}
	if msg.Params.RewardQuote.Asset != params.RewardQuote.Asset {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "quote asset change is prohibited")
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
