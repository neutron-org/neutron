package keeper_test

import (
	"fmt"
	"testing"
	"time"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/app/params"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	feerefundertypes "github.com/neutron-org/neutron/v6/x/feerefunder/types"
	"github.com/neutron-org/neutron/v6/x/interchaintxs/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/neutron-org/neutron/v6/testutil"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/interchaintxs/keeper"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/interchaintxs/types"
	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

const TestFeeCollectorAddr = "neutron1dua3d89szsmd3vwg0y5a2689ah0g4x68ps8vew"

const channelID = "channel-0"

var portID = "icacontroller-" + testutil.TestOwnerAddress + ICAId

func TestMsgRegisterInterchainAccountValidate(t *testing.T) {
	icak, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})

	tests := []struct {
		name        string
		msg         types.MsgRegisterInterchainAccount
		expectedErr error
	}{
		{
			"empty connection id",
			types.MsgRegisterInterchainAccount{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "",
				InterchainAccountId: "1",
			},
			types.ErrEmptyConnectionID,
		},
		{
			"empty fromAddress",
			types.MsgRegisterInterchainAccount{
				FromAddress:         "",
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid fromAddress",
			types.MsgRegisterInterchainAccount{
				FromAddress:         "invalid address",
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty interchain account id",
			types.MsgRegisterInterchainAccount{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "",
			},
			types.ErrEmptyInterchainAccountID,
		},
		{
			"long interchain account id",
			types.MsgRegisterInterchainAccount{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: string(make([]byte, 48)),
			},
			types.ErrLongInterchainAccountID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := icak.RegisterInterchainAccount(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestRegisterInterchainAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	icaMsgServer := mock_types.NewMockICAControllerMsgServer(ctrl)
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, wmKeeper, nil, icaKeeper, icaMsgServer, nil, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})

	msgRegAcc := types.MsgRegisterInterchainAccount{
		FromAddress:         testutil.TestOwnerAddress,
		ConnectionId:        "connection-0",
		InterchainAccountId: "ica0",
		Ordering:            channeltypes.ORDERED,
	}
	contractAddress := sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress)
	icaOwner := types.NewICAOwnerFromAddress(contractAddress, msgRegAcc.InterchainAccountId)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err := icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.ErrorContains(t, err, "is not a contract address")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	resp, err = icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to charge fees to pay for RegisterInterchainAccount msg")
	require.Nil(t, resp)

	msgRegAcc.RegisterFee = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1_000_000)))

	msgRegICA := &icacontrollertypes.MsgRegisterInterchainAccount{
		Owner:        icaOwner.String(),
		ConnectionId: msgRegAcc.ConnectionId,
		Version:      "",
		Ordering:     channeltypes.ORDERED,
	}

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee)
	icaMsgServer.EXPECT().RegisterInterchainAccount(ctx, msgRegICA).Return(nil, fmt.Errorf("failed to register ica"))
	resp, err = icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to RegisterInterchainAccount")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().
		SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee).
		Return(fmt.Errorf("failed to send coins"))
	resp, err = icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to charge fees to pay for RegisterInterchainAccount msg")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee)
	icaMsgServer.EXPECT().RegisterInterchainAccount(ctx, msgRegICA).Return(&icacontrollertypes.MsgRegisterInterchainAccountResponse{
		ChannelId: channelID,
		PortId:    portID,
	}, nil)
	icaKeeper.EXPECT().SetMiddlewareEnabled(ctx, portID, msgRegAcc.ConnectionId)
	resp, err = icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.NoError(t, err)
	require.Equal(t, types.MsgRegisterInterchainAccountResponse{
		ChannelId: channelID,
		PortId:    portID,
	}, *resp)
}

func TestRegisterInterchainAccountUnordered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	icaMsgServer := mock_types.NewMockICAControllerMsgServer(ctrl)
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, wmKeeper, nil, icaKeeper, icaMsgServer, nil, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})

	msgRegAcc := types.MsgRegisterInterchainAccount{
		FromAddress:         testutil.TestOwnerAddress,
		ConnectionId:        "connection-0",
		InterchainAccountId: "ica0",
		Ordering:            channeltypes.UNORDERED, // return unordered
	}
	contractAddress := sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress)
	icaOwner := types.NewICAOwnerFromAddress(contractAddress, msgRegAcc.InterchainAccountId)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err := icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.ErrorContains(t, err, "is not a contract address")
	require.Nil(t, resp)

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	resp, err = icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.ErrorContains(t, err, "failed to charge fees to pay for RegisterInterchainAccount msg")
	require.Nil(t, resp)

	msgRegAcc.RegisterFee = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1_000_000)))

	msgRegICA := &icacontrollertypes.MsgRegisterInterchainAccount{
		Owner:        icaOwner.String(),
		ConnectionId: msgRegAcc.ConnectionId,
		Version:      "",
		Ordering:     channeltypes.UNORDERED,
	}

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	wmKeeper.EXPECT().GetContractInfo(ctx, contractAddress).Return(&wasmtypes.ContractInfo{CodeID: 1})
	bankKeeper.EXPECT().SendCoins(ctx, sdk.MustAccAddressFromBech32(msgRegAcc.FromAddress), sdk.MustAccAddressFromBech32(TestFeeCollectorAddr), msgRegAcc.RegisterFee)
	icaMsgServer.EXPECT().RegisterInterchainAccount(ctx, msgRegICA).Return(&icacontrollertypes.MsgRegisterInterchainAccountResponse{
		ChannelId: channelID,
		PortId:    portID,
	}, nil)
	icaKeeper.EXPECT().SetMiddlewareEnabled(ctx, portID, msgRegAcc.ConnectionId)
	resp, err = icak.RegisterInterchainAccount(ctx, &msgRegAcc)
	require.NoError(t, err)
	require.Equal(t, types.MsgRegisterInterchainAccountResponse{
		ChannelId: channelID,
		PortId:    portID,
	}, *resp)
}

func TestMsgSubmitTXValidate(t *testing.T) {
	icak, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})

	cosmosMsg := codectypes.Any{
		TypeUrl: "msg",
		Value:   []byte{100}, // just check that values are not nil
	}

	tests := []struct {
		name        string
		msg         types.MsgSubmitTx
		expectedErr error
	}{
		{
			"invalid ack fee",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee: nil,
					AckFee: sdk.Coins{
						{
							Denom:  "{}!@#a",
							Amount: math.NewInt(100),
						},
					},
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"invalid timeout fee",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee: nil,
					AckFee:  sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.Coins{
						{
							Denom:  params.DefaultDenom,
							Amount: math.NewInt(-100),
						},
					},
				},
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"non-zero recv fee",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero ack fee",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     nil,
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero timeout fee",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: nil,
				},
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"empty connection id",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			types.ErrEmptyConnectionID,
		},
		{
			"empty FromAddress",
			types.MsgSubmitTx{
				FromAddress:         "",
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid FromAddress",
			types.MsgSubmitTx{
				FromAddress:         "invalid_address",
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty interchain account id",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			types.ErrEmptyInterchainAccountID,
		},
		{
			"no messages",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                nil,
				Timeout:             1,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			types.ErrNoMessages,
		},
		{
			"invalid timeout",
			types.MsgSubmitTx{
				FromAddress:         testutil.TestOwnerAddress,
				ConnectionId:        "connection-id",
				InterchainAccountId: "1",
				Msgs:                []*codectypes.Any{&cosmosMsg},
				Timeout:             0,
				Fee: feerefundertypes.Fee{
					RecvFee:    nil,
					AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
					TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
				},
			},
			types.ErrInvalidTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := icak.SubmitTx(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestSubmitTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	icaMsgServer := mock_types.NewMockICAControllerMsgServer(ctrl)
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	refundKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	channelKeeper := mock_types.NewMockChannelKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, wmKeeper, refundKeeper, icaKeeper, icaMsgServer, channelKeeper, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})

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
		Fee: feerefundertypes.Fee{
			RecvFee:    sdk.NewCoins(),
			AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
			TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(100))),
		},
	}

	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	icaOwner := types.NewICAOwnerFromAddress(contractAddress, submitMsg.InterchainAccountId)

	resp, err := icak.SubmitTx(ctx, nil)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "nil msg is prohibited")

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(false)
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address")

	params := icak.GetParams(ctx)
	maxMsgs := params.GetMsgSubmitTxMaxMessages()
	submitMsg.Msgs = make([]*codectypes.Any, maxMsgs+1)
	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "MsgSubmitTx contains more messages than allowed")
	submitMsg.Msgs = []*codectypes.Any{&cosmosMsg}

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return("", false)
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to GetActiveChannelID for port")

	activeChannel := "channel-0"
	// wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	// icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	// currCodec := icak.Codec
	// icak.Codec = &codec.AminoCodec{}
	// resp, err = icak.SubmitTx(ctx, &submitMsg)
	// icak.Codec = currCodec
	// require.Nil(t, resp)
	// require.ErrorContains(t, err, "only ProtoCodec is supported for receiving messages on the host chain")

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(uint64(0), false)
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "sequence send not found")

	sequence := uint64(100)
	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(fmt.Errorf("failed to lock fees"))
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to lock fees to pay for SubmitTx msg")

	data, err := keeper.SerializeCosmosTx(icak.Codec, submitMsg.Msgs)
	require.NoError(t, err)
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: submitMsg.Memo,
	}

	msgSendTx := &icacontrollertypes.MsgSendTx{
		Owner:           icaOwner.String(),
		ConnectionId:    submitMsg.ConnectionId,
		PacketData:      packetData,
		RelativeTimeout: uint64(time.Duration(submitMsg.Timeout) * time.Second), //nolint:gosec
	}

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(nil)
	icaMsgServer.EXPECT().SendTx(ctx, msgSendTx).Return(nil, fmt.Errorf("failed to send tx"))
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "failed to SendTx")

	wmKeeper.EXPECT().HasContractInfo(ctx, contractAddress).Return(true)
	icaKeeper.EXPECT().GetActiveChannelID(ctx, "connection-0", portID).Return(activeChannel, true)
	channelKeeper.EXPECT().GetNextSequenceSend(ctx, portID, activeChannel).Return(sequence, true)
	refundKeeper.EXPECT().LockFees(ctx, contractAddress, feerefundertypes.NewPacketID(portID, activeChannel, sequence), submitMsg.Fee).Return(nil)
	icaMsgServer.EXPECT().SendTx(ctx, msgSendTx).Return(&icacontrollertypes.MsgSendTxResponse{Sequence: sequence}, nil)
	resp, err = icak.SubmitTx(ctx, &submitMsg)
	require.Equal(t, types.MsgSubmitTxResponse{
		SequenceId: sequence,
		Channel:    activeChannel,
	}, *resp)
	require.NoError(t, err)
}

func TestMsgUpdateParamsValidate(t *testing.T) {
	icak, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})

	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectedErr string
	}{
		{
			"empty authority",
			types.MsgUpdateParams{
				Authority: "",
			},
			"authority is invalid",
		},
		{
			"invalid authority",
			types.MsgUpdateParams{
				Authority: "invalid authority",
			},
			"authority is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := icak.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
