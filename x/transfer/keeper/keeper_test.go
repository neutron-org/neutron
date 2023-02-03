package transfer_test

import (
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/app/params"
	"github.com/neutron-org/neutron/testutil"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/transfer/types"
)

const (
	TestAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

	reflectContractPath = "../../../wasmbinding/testdata/reflect.wasm"
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite KeeperTestSuite) TestTransfer() {
	suite.ConfigureTransferChannel()

	msgSrv := suite.GetNeutronZoneApp(suite.ChainA).TransferKeeper

	ctx := suite.ChainA.GetContext()
	resp, err := msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		Sender: "nonbech32",
	})
	suite.Nil(resp)
	suite.ErrorContains(err, "failed to parse address")

	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		SourcePort:    "transfer",
		SourceChannel: "nonexistent channel",
		Sender:        testutil.TestOwnerAddress,
	})
	suite.Nil(resp)
	suite.ErrorIs(err, channeltypes.ErrSequenceSendNotFound)

	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		SourcePort:    "nonexistent port",
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Sender:        testutil.TestOwnerAddress,
	})
	suite.Nil(resp)
	suite.ErrorIs(err, channeltypes.ErrSequenceSendNotFound)

	// sender is a non contract account
	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Sender:        testutil.TestOwnerAddress,
		Token:         sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(1000)),
	})
	suite.Nil(resp)
	suite.ErrorIs(err, errors.ErrInsufficientFunds)

	// sender is a non contract account
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(ctx, senderAddress, sdktypes.MustAccAddressFromBech32(testutil.TestOwnerAddress))
	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Sender:        testutil.TestOwnerAddress,
		Token:         sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(1000)),
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
	})
	suite.Equal(types.MsgTransferResponse{
		SequenceId: 1,
		Channel:    suite.TransferPath.EndpointA.ChannelID,
	}, *resp)
	suite.NoError(err)

	testOwner := sdktypes.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	// Store code and instantiate reflect contract.
	codeId := suite.StoreReflectCode(ctx, testOwner, reflectContractPath)
	contractAddress := suite.InstantiateReflectContract(ctx, testOwner, codeId)
	suite.Require().NotEmpty(contractAddress)

	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Sender:        contractAddress.String(),
		Token:         sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(1000)),
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
	})
	suite.Nil(resp)
	suite.ErrorContains(err, "failed to lock fees")

	suite.TopUpWallet(ctx, senderAddress, contractAddress)
	ctx = suite.ChainA.GetContext()
	resp, err = msgSrv.Transfer(sdktypes.WrapSDKContext(ctx), &types.MsgTransfer{
		SourcePort:    suite.TransferPath.EndpointA.ChannelConfig.PortID,
		SourceChannel: suite.TransferPath.EndpointA.ChannelID,
		Sender:        contractAddress.String(),
		Token:         sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(1000)),
		Receiver:      TestAddress,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 10,
			RevisionHeight: 10000,
		},
		Fee: feetypes.Fee{
			RecvFee:    nil,
			AckFee:     sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(1000))),
			TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(1000))),
		},
	})
	suite.Equal(types.MsgTransferResponse{
		SequenceId: 2,
		Channel:    suite.TransferPath.EndpointA.ChannelID,
	}, *resp)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TopUpWallet(ctx sdktypes.Context, sender sdktypes.AccAddress, contractAddress sdktypes.AccAddress) {
	coinsAmnt := sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(int64(1_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
