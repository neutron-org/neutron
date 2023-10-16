package types

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
)

const (
	TypeMsgCreateGauge    = "create_gauge"
	TypeMsgAddToGauge     = "add_to_gauge"
	TypeMsgStakeTokens    = "stake_tokens"
	TypeMsgBeginUnstaking = "begin_unstaking"
	TypeMsgUpdateParams   = "update-params"
)

var _ sdk.Msg = &MsgCreateGauge{}

// NewMsgCreateGauge creates a message to create a gauge with the provided parameters.
func NewMsgCreateGauge(
	isPerpetual bool,
	owner sdk.AccAddress,
	distributeTo QueryCondition,
	coins sdk.Coins,
	startTime time.Time,
	numEpochsPaidOver uint64,
	pricingTick int64,
) *MsgCreateGauge {
	return &MsgCreateGauge{
		IsPerpetual:       isPerpetual,
		Owner:             owner.String(),
		DistributeTo:      distributeTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		PricingTick:       pricingTick,
	}
}

// Route takes a create gauge message, then returns the RouterKey used for slashing.
func (m MsgCreateGauge) Route() string { return RouterKey }

// Type takes a create gauge message, then returns a create gauge message type.
func (m MsgCreateGauge) Type() string { return TypeMsgCreateGauge }

// ValidateBasic checks that the create gauge message is valid.
func (m MsgCreateGauge) ValidateBasic() error {
	if m.Owner == "" {
		return sdkerrors.Wrapf(ErrInvalidRequest, "owner should be set")
	}
	// TODO: If this is not set, infer start time as "now"
	if m.StartTime.Equal(time.Time{}) {
		return sdkerrors.Wrapf(ErrInvalidRequest, "distribution start time should be set")
	}
	if m.NumEpochsPaidOver == 0 {
		return sdkerrors.Wrapf(
			ErrInvalidRequest,
			"distribution period should be at least 1 epoch",
		)
	}
	if m.IsPerpetual && m.NumEpochsPaidOver != 1 {
		return sdkerrors.Wrapf(
			ErrInvalidRequest,
			"distribution period should be 1 epoch for perpetual gauge",
		)
	}
	if dextypes.IsTickOutOfRange(m.PricingTick) {
		return sdkerrors.Wrapf(
			ErrInvalidRequest,
			"pricing tick is out of range, must be between %d and %d",
			int64(dextypes.MaxTickExp)*-1,
			dextypes.MaxTickExp,
		)
	}
	if dextypes.IsTickOutOfRange(m.DistributeTo.StartTick) {
		return sdkerrors.Wrapf(
			ErrInvalidRequest,
			"start tick is out of range, must be between %d and %d",
			int64(dextypes.MaxTickExp)*-1,
			dextypes.MaxTickExp,
		)
	}
	if dextypes.IsTickOutOfRange(m.DistributeTo.EndTick) {
		return sdkerrors.Wrapf(
			ErrInvalidRequest,
			"start tick is out of range, must be between %d and %d",
			int64(dextypes.MaxTickExp)*-1,
			dextypes.MaxTickExp,
		)
	}

	return nil
}

// GetSignBytes takes a create gauge message and turns it into a byte array.
func (m MsgCreateGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes a create gauge message and returns the owner in a byte array.
func (m MsgCreateGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgAddToGauge{}

// NewMsgAddToGauge creates a message to add rewards to a specific gauge.
func NewMsgAddToGauge(owner sdk.AccAddress, gaugeID uint64, rewards sdk.Coins) *MsgAddToGauge {
	return &MsgAddToGauge{
		Owner:   owner.String(),
		GaugeId: gaugeID,
		Rewards: rewards,
	}
}

// Route takes an add to gauge message, then returns the RouterKey used for slashing.
func (m MsgAddToGauge) Route() string { return RouterKey }

// Type takes an add to gauge message, then returns an add to gauge message type.
func (m MsgAddToGauge) Type() string { return TypeMsgAddToGauge }

// ValidateBasic checks that the add to gauge message is valid.
func (m MsgAddToGauge) ValidateBasic() error {
	if m.Owner == "" {
		return sdkerrors.Wrapf(ErrInvalidRequest, "owner should be set")
	}
	if m.Rewards.Empty() {
		return sdkerrors.Wrapf(
			ErrInvalidRequest,
			"additional rewards should not be empty",
		)
	}

	return nil
}

// GetSignBytes takes an add to gauge message and turns it into a byte array.
func (m MsgAddToGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes an add to gauge message and returns the owner in a byte array.
func (m MsgAddToGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgStake{}

// NewMsgStakeTokens creates a message to stake tokens.
func NewMsgSetupStake(owner sdk.AccAddress, duration time.Duration, coins sdk.Coins) *MsgStake {
	return &MsgStake{
		Owner: owner.String(),
		Coins: coins,
	}
}

func (m MsgStake) Route() string { return RouterKey }
func (m MsgStake) Type() string  { return TypeMsgStakeTokens }
func (m MsgStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "Invalid owner address (%s)", err)
	}

	if !m.Coins.IsAllPositive() {
		return fmt.Errorf("cannot stake up a zero or negative amount")
	}

	return nil
}

func (m MsgStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgStake) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgUnstake{}

func NewMsgUnstakeDescriptor(id uint64, coins sdk.Coins) *MsgUnstake_UnstakeDescriptor {
	return &MsgUnstake_UnstakeDescriptor{
		ID:    id,
		Coins: coins,
	}
}

// NewMsgUnstake creates a message to unstake the tokens of a set of stake records.
func NewMsgUnstake(owner sdk.AccAddress, unstakes []*MsgUnstake_UnstakeDescriptor) *MsgUnstake {
	return &MsgUnstake{
		Owner:    owner.String(),
		Unstakes: unstakes,
	}
}

func (m MsgUnstake) Route() string { return RouterKey }
func (m MsgUnstake) Type() string  { return TypeMsgBeginUnstaking }
func (m MsgUnstake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "Invalid owner address (%s)", err)
	}

	for _, unstake := range m.Unstakes {
		if unstake.ID == 0 {
			return fmt.Errorf("invalid stake ID, got %v", unstake.ID)
		}

		if !unstake.Coins.Empty() && !unstake.Coins.IsAllPositive() {
			return fmt.Errorf("cannot unstake a zero or negative amount")
		}
	}

	return nil
}

func (m MsgUnstake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgUnstake) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgUpdateParams) Route() string { return RouterKey }

func (msg *MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "authority is invalid")
	}
	if err := msg.Params.Validate(); err != nil {
		return err
	}
	return nil
}
