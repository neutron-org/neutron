package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/neutron-org/neutron/v7/testutil"
	testutil_keeper "github.com/neutron-org/neutron/v7/testutil/feerefunder/keeper"
	mock_types "github.com/neutron-org/neutron/v7/testutil/mocks/feerefunder/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v7/x/feerefunder/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

const (
	TestAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"
)

func TestKeeperCheckFees(t *testing.T) {
	k, ctx := testutil_keeper.FeeKeeper(t, nil, nil)

	err := k.SetParams(ctx, types.Params{
		MinFee: types.Fee{
			RecvFee:    nil,
			AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(100)), sdk.NewCoin("denom2", math.NewInt(100))),
			TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(100)), sdk.NewCoin("denom2", math.NewInt(100))),
		},
		FeeEnabled: true,
	})
	require.NoError(t, err)

	for _, tc := range []struct {
		desc    string
		fees    *types.Fee
		minFees types.Fee
		err     error
	}{
		{
			desc: "SingleProperDenomInsufficient",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
			},
			err: sdkerrors.ErrInsufficientFee,
		},
		{
			desc: "SufficientTimeout-InsufficientAck",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
			},
			err: sdkerrors.ErrInsufficientFee,
		},
		{
			desc: "NonNilRecvFee",
			fees: &types.Fee{
				RecvFee:    sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
			},
			err: fmt.Errorf("recv fee must be zero: %w", sdkerrors.ErrInvalidCoins),
		},
		{
			desc: "SingleDenomSufficient",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
			},
			err: nil,
		},
		{
			desc: "MultipleDenomsOneIsEnough",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101)), sdk.NewCoin("denom2", math.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101)), sdk.NewCoin("denom2", math.NewInt(1))),
			},
			err: nil,
		},
		{
			desc: "NoProperDenom",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom3", math.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom3", math.NewInt(1))),
			},
			err: sdkerrors.ErrInsufficientFee,
		},
		{
			desc: "ProperDenomPlusRandomAckOne",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101)), sdk.NewCoin("denom3", math.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			desc: "ProperDenomPlusRandomTimeoutOne",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101)), sdk.NewCoin("denom3", math.NewInt(1))),
			},
			err: sdkerrors.ErrInvalidCoins,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := k.CheckFees(ctx, *tc.fees)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), errors.Unwrap(err).Error())
			}
		})
	}
}

func TestKeeperLockFees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	channelKeeper := mock_types.NewMockChannelKeeper(ctrl)
	k, ctx := testutil_keeper.FeeKeeper(t, channelKeeper, bankKeeper)

	payer := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	err := k.SetParams(ctx, types.Params{
		MinFee: types.Fee{
			RecvFee:    nil,
			AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(100)), sdk.NewCoin("denom2", math.NewInt(100))),
			TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(100)), sdk.NewCoin("denom2", math.NewInt(100))),
		},
		FeeEnabled: true,
	})
	require.NoError(t, err)

	packet := types.PacketID{
		ChannelId: "channel-0",
		PortId:    "transfer",
		Sequence:  111,
	}

	// channel not found
	channelKeeper.EXPECT().GetChannel(ctx, packet.PortId, packet.ChannelId).Return(channeltypes.Channel{}, false)
	err = k.LockFees(ctx, payer, packet, types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(100))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(100))),
	})
	require.True(t, channeltypes.ErrChannelNotFound.Is(err))

	// bank send error
	validFee := types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(101))),
	}
	channelKeeper.EXPECT().GetChannel(ctx, packet.PortId, packet.ChannelId).Return(channeltypes.Channel{}, true)
	bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, validFee.Total()).Return(fmt.Errorf("bank error"))
	err = k.LockFees(ctx, payer, packet, validFee)
	require.ErrorContains(t, err, "bank error")

	//  valid case
	channelKeeper.EXPECT().GetChannel(ctx, packet.PortId, packet.ChannelId).Return(channeltypes.Channel{}, true)
	bankKeeper.EXPECT().SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, validFee.Total()).Return(nil)
	err = k.LockFees(ctx, payer, packet, validFee)
	require.NoError(t, err)
	require.Equal(t, sdk.Events{
		sdk.NewEvent(
			types.EventTypeLockFees,
			sdk.NewAttribute(types.AttributeKeyPayer, payer.String()),
			sdk.NewAttribute(types.AttributeKeyPortID, packet.PortId),
			sdk.NewAttribute(types.AttributeKeyChannelID, packet.ChannelId),
			sdk.NewAttribute(types.AttributeKeySequence, strconv.FormatUint(packet.Sequence, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	}, ctx.EventManager().Events())

	// insufficient fee
	err = k.LockFees(ctx, payer, packet, types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
	})
	require.True(t, sdkerrors.ErrInsufficientFee.Is(err))

	// ack fee is empty
	err = k.LockFees(ctx, payer, packet, types.Fee{
		RecvFee:    nil,
		AckFee:     nil,
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
	})
	require.ErrorContains(t, err, "ack fee or timeout fee is zero")

	// timeout fee is empty
	err = k.LockFees(ctx, payer, packet, types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", math.NewInt(1))),
		TimeoutFee: nil,
	})
	require.ErrorContains(t, err, "ack fee or timeout fee is zero")
}

func TestDistributeAcknowledgementFee(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	channelKeeper := mock_types.NewMockChannelKeeper(ctrl)
	k, ctx := testutil_keeper.FeeKeeper(t, channelKeeper, bankKeeper)

	validFee := types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(1001))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(2001))),
	}
	packet := types.PacketID{
		ChannelId: "channel-0",
		PortId:    "transfer",
		Sequence:  111,
	}
	payer := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	receiver := sdk.MustAccAddressFromBech32(TestAddress)

	// prepare the fees we want to distribute
	k.StoreFeeInfo(ctx, types.FeeInfo{
		Payer:    payer.String(),
		Fee:      validFee,
		PacketId: packet,
	})

	panicErrorToCatch := errors.Wrapf(errors.Wrapf(fmt.Errorf("bank module error"), "error distributing fee to a receiver: %s", receiver.String()), "error distributing ack fee: receiver = %s, packetID=%v", receiver, packet)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, validFee.AckFee).Return(fmt.Errorf("bank module error"))
	assert.PanicsWithError(t, panicErrorToCatch.Error(), func() { k.DistributeAcknowledgementFee(ctx, receiver, packet) })

	panicErrorToCatch = errors.Wrapf(errors.Wrapf(fmt.Errorf("bank module error"), "error distributing fee to a receiver: %s", payer.String()), "error distributing unused timeout fee: receiver = %s, packetID=%v", receiver, packet)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, validFee.AckFee).Return(nil)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, payer, validFee.TimeoutFee).Return(fmt.Errorf("bank module error"))
	assert.PanicsWithError(t, panicErrorToCatch.Error(), func() { k.DistributeAcknowledgementFee(ctx, receiver, packet) })

	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, validFee.AckFee).Return(nil)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, payer, validFee.TimeoutFee).Return(nil)
	assert.NotPanics(t, func() { k.DistributeAcknowledgementFee(ctx, receiver, packet) })
	require.Equal(t, sdk.Events{
		sdk.NewEvent(
			types.EventTypeDistributeAcknowledgementFee,
			sdk.NewAttribute(types.AttributeKeyReceiver, TestAddress),
			sdk.NewAttribute(types.AttributeKeyPortID, packet.PortId),
			sdk.NewAttribute(types.AttributeKeyChannelID, packet.ChannelId),
			sdk.NewAttribute(types.AttributeKeySequence, strconv.FormatUint(packet.Sequence, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	}, ctx.EventManager().Events())

	feeInfo, found := k.GetFeeInfo(ctx, packet)
	require.Nil(t, feeInfo)
	require.False(t, found, "no expected fee info")

	// simulate fee disabled
	noFeePacket := types.PacketID{
		ChannelId: "channel-0",
		PortId:    "transfer",
		Sequence:  112123,
	}
	// does not call fee transfers
	assert.NotPanics(t, func() { k.DistributeAcknowledgementFee(ctx, receiver, noFeePacket) })
}

func TestDistributeTimeoutFee(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	channelKeeper := mock_types.NewMockChannelKeeper(ctrl)
	k, ctx := testutil_keeper.FeeKeeper(t, channelKeeper, bankKeeper)

	validFee := types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(1001))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(2001))),
	}
	packet := types.PacketID{
		ChannelId: "channel-0",
		PortId:    "transfer",
		Sequence:  111,
	}
	payer := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	receiver := sdk.MustAccAddressFromBech32(TestAddress)

	// prepare the fees we want to distribute
	k.StoreFeeInfo(ctx, types.FeeInfo{
		Payer:    payer.String(),
		Fee:      validFee,
		PacketId: packet,
	})

	panicErrorToCatch := errors.Wrapf(errors.Wrapf(fmt.Errorf("bank module error"), "error distributing fee to a receiver: %s", receiver.String()), "error distributing timeout fee: receiver = %s, packetID=%v", receiver, packet)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, validFee.TimeoutFee).Return(fmt.Errorf("bank module error"))
	assert.PanicsWithError(t, panicErrorToCatch.Error(), func() { k.DistributeTimeoutFee(ctx, receiver, packet) })

	panicErrorToCatch = errors.Wrapf(errors.Wrapf(fmt.Errorf("bank module error"), "error distributing fee to a receiver: %s", payer.String()), "error distributing unused ack fee: receiver = %s, packetID=%v", receiver, packet)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, validFee.TimeoutFee).Return(nil)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, payer, validFee.AckFee).Return(fmt.Errorf("bank module error"))
	assert.PanicsWithError(t, panicErrorToCatch.Error(), func() { k.DistributeTimeoutFee(ctx, receiver, packet) })

	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, validFee.TimeoutFee).Return(nil)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, payer, validFee.AckFee).Return(nil)
	assert.NotPanics(t, func() { k.DistributeTimeoutFee(ctx, receiver, packet) })
	require.Equal(t, sdk.Events{
		sdk.NewEvent(
			types.EventTypeDistributeTimeoutFee,
			sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
			sdk.NewAttribute(types.AttributeKeyPortID, packet.PortId),
			sdk.NewAttribute(types.AttributeKeyChannelID, packet.ChannelId),
			sdk.NewAttribute(types.AttributeKeySequence, strconv.FormatUint(packet.Sequence, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	}, ctx.EventManager().Events())

	feeInfo, found := k.GetFeeInfo(ctx, packet)
	require.Nil(t, feeInfo)
	require.False(t, found, "no expected fee info")

	// simulate fee disabled
	noFeePacket := types.PacketID{
		ChannelId: "channel-0",
		PortId:    "transfer",
		Sequence:  112123,
	}
	// does not call fee transfers
	assert.NotPanics(t, func() { k.DistributeTimeoutFee(ctx, receiver, noFeePacket) })
}

func TestFeeInfo(t *testing.T) {
	k, ctx := testutil_keeper.FeeKeeper(t, nil, nil)
	validFee := types.Fee{
		RecvFee:    nil,
		AckFee:     sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(1001))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(1001))),
	}
	for i := uint64(0); i < 1000; i++ {
		packet := types.PacketID{
			ChannelId: "channel-0",
			PortId:    "transfer",
			Sequence:  i,
		}
		payer := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

		k.StoreFeeInfo(ctx, types.FeeInfo{
			Payer:    payer.String(),
			Fee:      validFee,
			PacketId: packet,
		})
	}
	infos := k.GetAllFeeInfos(ctx)
	require.Equal(t, 1000, len(infos))
}
