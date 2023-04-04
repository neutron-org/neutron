package types_test

import (
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/stretchr/testify/require"

	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/transfer/types"
)

const TestAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

func TestMsgSubmitTXValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		// We can check only fee validity because we didn't change original ValidateBasic call
		{
			"valid",
			func() sdktypes.Msg {
				return &types.MsgTransfer{
					SourcePort:    "port_id",
					SourceChannel: "channel_id",
					Token:         sdktypes.NewCoin("denom", sdktypes.NewInt(100)),
					Sender:        TestAddress,
					Receiver:      TestAddress,
					TimeoutHeight: ibcclienttypes.Height{
						RevisionNumber: 100,
						RevisionHeight: 100,
					},
					TimeoutTimestamp: 10000,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
					},
				}
			},
			nil,
		},
		{
			"non-zero recv fee",
			func() sdktypes.Msg {
				return &types.MsgTransfer{
					SourcePort:    "port_id",
					SourceChannel: "channel_id",
					Token:         sdktypes.NewCoin("denom", sdktypes.NewInt(100)),
					Sender:        TestAddress,
					Receiver:      TestAddress,
					TimeoutHeight: ibcclienttypes.Height{
						RevisionNumber: 100,
						RevisionHeight: 100,
					},
					TimeoutTimestamp: 10000,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
					},
				}
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero ack fee",
			func() sdktypes.Msg {
				return &types.MsgTransfer{
					SourcePort:    "port_id",
					SourceChannel: "channel_id",
					Token:         sdktypes.NewCoin("denom", sdktypes.NewInt(100)),
					Sender:        TestAddress,
					Receiver:      TestAddress,
					TimeoutHeight: ibcclienttypes.Height{
						RevisionNumber: 100,
						RevisionHeight: 100,
					},
					TimeoutTimestamp: 10000,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
					},
				}
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero timeout fee",
			func() sdktypes.Msg {
				return &types.MsgTransfer{
					SourcePort:    "port_id",
					SourceChannel: "channel_id",
					Token:         sdktypes.NewCoin("denom", sdktypes.NewInt(100)),
					Sender:        TestAddress,
					Receiver:      TestAddress,
					TimeoutHeight: ibcclienttypes.Height{
						RevisionNumber: 100,
						RevisionHeight: 100,
					},
					TimeoutTimestamp: 10000,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", sdktypes.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(),
					},
				}
			},
			sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}
