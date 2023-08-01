package types

import (
	"strings"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewPacketID(portID, channelID string, sequence uint64) PacketID {
	return PacketID{
		ChannelId: channelID,
		PortId:    portID,
		Sequence:  sequence,
	}
}

// NewFee creates and returns a new Fee struct encapsulating the receive, acknowledgement and timeout fees as sdk.Coins
func NewFee(recvFee, ackFee, timeoutFee sdk.Coins) Fee {
	return Fee{
		RecvFee:    recvFee,
		AckFee:     ackFee,
		TimeoutFee: timeoutFee,
	}
}

// Total returns the total amount for a given Fee
func (m Fee) Total() sdk.Coins {
	return m.RecvFee.Add(m.AckFee...).Add(m.TimeoutFee...)
}

// Validate asserts that each Fee is valid:
// * RecvFee must be zero;
// * AckFee and TimeoutFee must be non-zero
func (m Fee) Validate() error {
	var errFees []string
	if !m.AckFee.IsValid() {
		errFees = append(errFees, "ack fee is invalid")
	}
	if !m.TimeoutFee.IsValid() {
		errFees = append(errFees, "timeout fee invalid")
	}

	if len(errFees) > 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "contains invalid fees: %s", strings.Join(errFees, " , "))
	}

	if !m.RecvFee.IsZero() {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "recv fee must be zero")
	}

	// if ack or timeout fees are zero or empty return an error
	if m.AckFee.IsZero() || m.TimeoutFee.IsZero() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "ack fee or timeout fee is zero")
	}

	return nil
}
