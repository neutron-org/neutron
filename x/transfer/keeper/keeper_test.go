package transfer_test

import (
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcerrors "github.com/cosmos/ibc-go/v8/modules/core/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/app/params"
	"github.com/neutron-org/neutron/v6/testutil"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/transfer/types"
	"github.com/neutron-org/neutron/v6/testutil/transfer/keeper"
	feetypes "github.com/neutron-org/neutron/v6/x/feerefunder/types"
	"github.com/neutron-org/neutron/v6/x/transfer/types"
)

const (
	TestAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

	reflectContractPath = "../../../wasmbinding/testdata/reflect.wasm"
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite KeeperTestSuite) TestTransfer() { //nolint:govet // it's a test so it's okay to copy locks
	suite.ConfigureTransferChannel()

	msgSrv := suite.GetNeutronZoneApp(suite.ChainA).TransferKeeper

	ctx := suite.ChainA.GetContext()
	resp, err := msgSrv.Transfer(ctx, &types.MsgTransfer{
		Sender: "nonbech32",
	})
	suite.Nil(resp)
	suite.ErrorIs(err, errors.ErrInvalidAddress)

	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(ctx, &types.MsgTransfer{
		SourcePort:    "nonexistent_port",
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000)),
		Sender:        testutil.TestOwnerAddress,
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
		Fee: feetypes.Fee{
			RecvFee:    nil,
			AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
			TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
		},
	})
	suite.Nil(resp)
	suite.ErrorIs(err, channeltypes.ErrSequenceSendNotFound)

	// sender is a non contract account
	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(ctx, &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000)),
		Sender:        testutil.TestOwnerAddress,
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
		Fee: feetypes.Fee{
			RecvFee:    nil,
			AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
			TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
		},
	})
	suite.Nil(resp)
	suite.ErrorIs(err, errors.ErrInsufficientFunds)

	// sender is a non contract account
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(ctx, senderAddress, sdktypes.MustAccAddressFromBech32(testutil.TestOwnerAddress))
	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(ctx, &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000)),
		Sender:        testutil.TestOwnerAddress,
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
		Fee: feetypes.Fee{
			RecvFee:    nil,
			AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
			TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
		},
	})
	suite.Equal(types.MsgTransferResponse{
		SequenceId: 1,
		Channel:    suite.TransferPath.EndpointA.ChannelID,
	}, *resp)
	suite.NoError(err)

	testOwner := sdktypes.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	// Store code and instantiate reflect contract.
	codeID := suite.StoreTestCode(ctx, testOwner, reflectContractPath)
	contractAddress := suite.InstantiateTestContract(ctx, testOwner, codeID)
	suite.Require().NotEmpty(contractAddress)

	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(ctx, &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000)),
		Sender:        contractAddress.String(),
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
		Fee: feetypes.Fee{
			RecvFee:    nil,
			AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
			TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
		},
	})
	suite.Nil(resp)
	suite.ErrorContains(err, "failed to lock fees")

	suite.TopUpWallet(ctx, senderAddress, contractAddress)
	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(ctx, &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000)),
		Sender:        contractAddress.String(),
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
		Fee: feetypes.Fee{
			RecvFee:    nil,
			AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
			TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(1000))),
		},
	})
	suite.Equal(types.MsgTransferResponse{
		SequenceId: 2,
		Channel:    suite.TransferPath.EndpointA.ChannelID,
	}, *resp)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TopUpWallet(ctx sdktypes.Context, sender, contractAddress sdktypes.AccAddress) {
	coinsAmnt := sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func TestMsgTransferValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(transfertypes.ModuleName).Return([]byte("address"))
	k, ctx, _ := keeper.TransferKeeper(t, wmKeeper, nil, nil, authKeeper)

	wmKeeper.EXPECT().HasContractInfo(ctx, gomock.Any()).Return(true).AnyTimes()

	tests := []struct {
		name        string
		msg         types.MsgTransfer
		expectedErr error
	}{
		{
			"invalid ack fee",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee: nil,
					AckFee: sdktypes.Coins{
						{
							Denom:  "{}!@#a",
							Amount: math.NewInt(100),
						},
					},
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			errors.ErrInvalidCoins,
		},
		{
			"invalid timeout fee",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee: nil,
					AckFee:  sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.Coins{
						{
							Denom:  params.DefaultDenom,
							Amount: math.NewInt(-100),
						},
					},
				},
			},
			errors.ErrInvalidCoins,
		},
		{
			"non-zero recv fee",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			errors.ErrInvalidCoins,
		},
		{
			"zero ack fee",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     nil,
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			errors.ErrInvalidCoins,
		},
		{
			"zero timeout fee",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: nil,
				},
			},
			errors.ErrInvalidCoins,
		},
		{
			"empty source port",
			types.MsgTransfer{
				SourcePort:    "",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid source port separator",
			types.MsgTransfer{
				SourcePort:    "/transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid source port length",
			types.MsgTransfer{
				SourcePort:    "t",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid source port",
			types.MsgTransfer{
				SourcePort:    "nonexistent port",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"empty source channel",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid source channel separator",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "/channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid source channel length",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: string(make([]byte, host.DefaultMaxCharacterLength+1)),
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid source channel",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel 2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			host.ErrInvalidID,
		},
		{
			"invalid token denom",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token: sdktypes.Coin{
					Denom:  "{}!@#a",
					Amount: math.NewInt(100),
				},
				Sender:   testutil.TestOwnerAddress,
				Receiver: TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			ibcerrors.ErrInvalidCoins,
		},
		{
			"nil token amount",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token: sdktypes.Coin{
					Denom: params.DefaultDenom,
				},
				Sender:   testutil.TestOwnerAddress,
				Receiver: TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			ibcerrors.ErrInvalidCoins,
		},
		{
			"negative token amount",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token: sdktypes.Coin{
					Denom:  params.DefaultDenom,
					Amount: math.NewInt(-100),
				},
				Sender:   testutil.TestOwnerAddress,
				Receiver: TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			ibcerrors.ErrInvalidCoins,
		},
		{
			"empty sender",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        "",
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			errors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        "invalid_sender",
				Receiver:      TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			errors.ErrInvalidAddress,
		},
		{
			"empty receiver",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      "",
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			ibcerrors.ErrInvalidAddress,
		},
		{
			"long receiver",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      string(make([]byte, transfertypes.MaximumReceiverLength+1)),
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			ibcerrors.ErrInvalidAddress,
		},
		{
			"long memo",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token:         sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				Sender:        testutil.TestOwnerAddress,
				Receiver:      TestAddress,
				Memo:          string(make([]byte, transfertypes.MaximumMemoLength+1)),
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			transfertypes.ErrInvalidMemo,
		},
		{
			"invalid token denom prefix format",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token: sdktypes.Coin{
					Denom:  transfertypes.DenomPrefix,
					Amount: math.NewInt(100),
				},
				Sender:   testutil.TestOwnerAddress,
				Receiver: TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			transfertypes.ErrInvalidDenomForTransfer,
		},
		{
			"invalid token denom prefix format with separator",
			types.MsgTransfer{
				SourcePort:    "transfer",
				SourceChannel: "channel-2",
				Token: sdktypes.Coin{
					Denom:  transfertypes.DenomPrefix + "/",
					Amount: math.NewInt(100),
				},
				Sender:   testutil.TestOwnerAddress,
				Receiver: TestAddress,
				Fee: feetypes.Fee{
					RecvFee:    nil,
					AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			transfertypes.ErrInvalidDenomForTransfer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := k.Transfer(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
