package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/neutron-org/neutron/v3/app/params"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	feerefundertypes "github.com/neutron-org/neutron/v3/x/feerefunder/types"
	"github.com/neutron-org/neutron/v3/x/interchaintxs/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/neutron-org/neutron/v3/testutil"
	testkeeper "github.com/neutron-org/neutron/v3/testutil/interchaintxs/keeper"
	mock_types "github.com/neutron-org/neutron/v3/testutil/mocks/interchaintxs/types"
	"github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

const TestFeeCollectorAddr = "neutron1dua3d89szsmd3vwg0y5a2689ah0g4x68ps8vew"

func TestRegisterInterchainAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, wmKeeper, nil, icaKeeper, nil, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})
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

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.ErrorContains(t, err, "is not a contract address")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to charge fees to pay for RegisterInterchainAccount msg")
	require.Nil(t, resp)

	msgRegAcc.RegisterFee = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(1_000_000)))

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee)
	icaKeeper.EXPECT().RegisterInterchainAccount(ctx, msgRegAcc.ConnectionId, icaOwner.String(), "").Return(fmt.Errorf("failed to register ica"))
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to RegisterInterchainAccount")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().
		SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee).
		Return(fmt.Errorf("failed to send coins"))
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to send coins")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee)
	icaKeeper.EXPECT().RegisterInterchainAccount(ctx, msgRegAcc.ConnectionId, icaOwner.String(), "").Return(nil)
	resp, err = icak.RegisterInterchainAccount(goCtx, &msgRegAcc)
	require.NoError(t, err)
	require.Equal(t, types.MsgRegisterInterchainAccountResponse{}, *resp)
}

func TestSubmitTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	refundKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	channelKeeper := mock_types.NewMockChannelKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, wmKeeper, refundKeeper, icaKeeper, channelKeeper, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})
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

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address")

	params := icak.GetParams(ctx)
	maxMsgs := params.GetMsgSubmitTxMaxMessages()
	submitMsg.Msgs = make([]*codectypes.Any, maxMsgs+1)
	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "MsgSubmitTx contains more messages than allowed")
	submitMsg.Msgs = []*codectypes.Any{&cosmosMsg}

	portID := "icacontroller-" + testutil.TestOwnerAddress + ICAId
	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return("", false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to GetActiveChannelID for port")

	activeChannel := "channel-0"
	// wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	// icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	// currCodec := icak.Codec
	// icak.Codec = &codec.AminoCodec{}
	// resp, err = icak.SubmitTx(goCtx, &submitMsg)
	// icak.Codec = currCodec
	// require.Nil(t, resp)
	// require.ErrorContains(t, err, "only ProtoCodec is supported for receiving messages on the host chain")

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(uint64(0), false)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "sequence send not found")

	sequence := uint64(100)
	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
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
	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(nil)
	icaKeeper.EXPECT().SendTx(ctx, nil, "connection-0", portID, packetData, uint64(timeoutTimestamp)).Return(uint64(0), fmt.Errorf("faile to send tx"))
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to SendTx")

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(nil)
	icaKeeper.EXPECT().SendTx(ctx, nil, "connection-0", portID, packetData, uint64(timeoutTimestamp)).Return(uint64(0), nil)
	resp, err = icak.SubmitTx(goCtx, &submitMsg)
	require.Equal(t, types.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    activeChannel,
	}, *resp)
	require.NoError(t, err)
}
