package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	types2 "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"

	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchaintxs/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/neutron-org/neutron/testutil"
	testkeeper "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/interchaintxs/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

func TestRegisterInterchainAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, cmKeeper, nil, icaKeeper, nil, nil)
	goCtx := sdk.WrapSDKContext(ctx)

	msgRegAcc := types.MsgRegisterInterchainAccount{
		FromAddress:         testutil.TestOwnerAddress,
		ConnectionId:        "connection-0",
		InterchainAccountId: "ica0",
	}
	contractAddress := sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress)
	icaOwner := types.NewICAOwnerFromAddress(contractAddress, msgRegAcc.InterchainAccountId)

	resp, err := icak.RegisterInterchainAccount(goCtx, &types.MsgRegisterInterchainAccount{})
	require.ErrorContains(t, err, "failed to parse address")
	require.Nil(t, resp)

	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.ErrorContains(t, err, "is not a contract address")
	require.Nil(t, resp)

	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().RegisterInterchainAccount(ctx, msgRegAcc.ConnectionId, icaOwner.String(), "").Return(fmt.Errorf("failed to register ica"))
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to RegisterInterchainAccount")
	require.Nil(t, resp)

	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().RegisterInterchainAccount(ctx, msgRegAcc.ConnectionId, icaOwner.String(), "").Return(nil)
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.NoError(t, err)
	require.Equal(t, types.MsgRegisterInterchainAccountResponse{}, *resp)
}

func TestSubmitTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	capabilityKeeper := mock_types.NewMockScopedKeeper(ctrl)
	refundKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	channelKeeper := mock_types.NewMockChannelKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, cmKeeper, refundKeeper, icaKeeper, channelKeeper, capabilityKeeper)
	goCtx := sdk.WrapSDKContext(ctx)

	cosmosMsg := codectypes.Any{
		TypeUrl: "/cosmos.staking.v1beta1.MsgDelegate",
		Value:   []byte{26, 10, 10, 5, 115, 116, 97, 107, 101, 18, 1, 48},
	}
	submitMsg := types.MsgSubmitTx{
		FromAddress:         testutil.TestOwnerAddress,
		InterchainAccountId: "ica0",
		ConnectionId:        "connection-0",
		Msgs:                []*codectypes.Any{&cosmosMsg},
		Memo:                "memo",
		Timeout:             100,
		Fee:                 feerefundertypes.Fee{},
	}

	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	resp, err := icak.SubmitTx(goCtx, nil)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "nil msg is prohibited")

	resp, err = icak.SubmitTx(goCtx, &types.MsgSubmitTx{})
	require.Nil(t, resp)
	require.ErrorContains(t, err, "empty Msgs field is prohibited")

	resp, err = icak.SubmitTx(goCtx, &types.MsgSubmitTx{Msgs: []*codectypes.Any{&cosmosMsg}})
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to parse address")

	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address")

	params := icak.GetParams(ctx)
	maxMsgs := params.GetMsgSubmitTxMaxMessages()
	submitMsg.Msgs = make([]*codectypes.Any, maxMsgs+1)
	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "MsgSubmitTx contains more messages than allowed")
	submitMsg.Msgs = []*codectypes.Any{&cosmosMsg}

	portID := "icacontroller-" + testutil.TestOwnerAddress + ".ica0"
	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return("", false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to GetActiveChannelID for port")

	activeChannel := "channel-0"
	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	capabilityKeeper.EXPECT().GetCapability(ctx, host.ChannelCapabilityPath(portID, activeChannel)).Return(nil, false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to GetCapability")

	capability := types2.Capability{Index: 1}
	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	capabilityKeeper.EXPECT().GetCapability(ctx, host.ChannelCapabilityPath(portID, activeChannel)).Return(&capability, true)
	currCodec := icak.Codec
	icak.Codec = &codec.AminoCodec{}
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	icak.Codec = currCodec
	require.Nil(t, resp)
	require.ErrorContains(t, err, "only ProtoCodec is supported for receiving messages on the host chain")

	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	capabilityKeeper.EXPECT().GetCapability(ctx, host.ChannelCapabilityPath(portID, activeChannel)).Return(&capability, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(uint64(0), false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "sequence send not found")

	sequence := uint64(100)
	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	capabilityKeeper.EXPECT().GetCapability(ctx, host.ChannelCapabilityPath(portID, activeChannel)).Return(&capability, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(fmt.Errorf("failed to lock fees"))
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to lock fees to pay for SubmitTx msg")

	data, err := keeper.SerializeCosmosTx(icak.Codec, submitMsg.Msgs)
	require.NoError(t, err)
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: submitMsg.Memo,
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Duration(submitMsg.Timeout) * time.Second).UnixNano()
	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	capabilityKeeper.EXPECT().GetCapability(ctx, host.ChannelCapabilityPath(portID, activeChannel)).Return(&capability, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(nil)
	icaKeeper.EXPECT().SendTx(ctx, &capability, "connection-0", portID, packetData, uint64(timeoutTimestamp)).Return(uint64(0), fmt.Errorf("faile to send tx"))
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to SendTx")

	cmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	capabilityKeeper.EXPECT().GetCapability(ctx, host.ChannelCapabilityPath(portID, activeChannel)).Return(&capability, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(nil)
	icaKeeper.EXPECT().SendTx(ctx, &capability, "connection-0", portID, packetData, uint64(timeoutTimestamp)).Return(uint64(0), nil)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Equal(t, types.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    activeChannel,
	}, *resp)
	require.NoError(t, err)
}
