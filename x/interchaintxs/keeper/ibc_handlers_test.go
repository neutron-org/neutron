package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/neutron-org/neutron/testutil"
	testkeeper "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/interchaintxs/types"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchaintxs/keeper"
	"github.com/stretchr/testify/require"
)

func TestHandleAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	icak, infCtx := testkeeper.InterchainTxsKeeper(t, cmKeeper, feeKeeper, icaKeeper, nil, nil)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))

	errACK := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: "error",
		},
	}
	errAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&errACK)
	require.NoError(t, err)
	resACK := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{Result: []byte("Result")},
	}
	resAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&resACK)
	require.NoError(t, err)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.PortPrefix + testutil.TestOwnerAddress + ".ica0",
		SourceChannel: "channel-0",
	}
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)

	err = icak.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	require.ErrorContains(t, err, "failed to get ica owner from port")

	err = icak.HandleAcknowledgement(ctx, p, nil, relayerAddress)
	require.ErrorContains(t, err, "cannot unmarshal ICS-27 packet acknowledgement")

	// error during SudoResponse
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Return(nil, fmt.Errorf("SudoResponse error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// error during SudoError
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	//  success during SudoError
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Return(nil, nil)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// out of gas during SudoError
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, error string) {
		ctx.GasMeter().ConsumeGas(ctx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	// feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// check we have ReserveGas reserved and
	// check gas consumption from cachedCtx has added to the main ctx
	// one of the ways to check it - make the check during SudoResponse call
	gasReserved := false
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+keeper.GasReserve {
			gasReserved = true
		}
		cachedCtx.GasMeter().ConsumeGas(1_000_000, "Sudo response consumption")
	}).Return(nil, nil)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	consumedBefore := ctx.GasMeter().GasConsumed()
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.True(t, gasReserved)
	require.Equal(t, consumedBefore+1_000_000, ctx.GasMeter().GasConsumed())

	// not enough gas to reserve + not enough to make AddContractFailure failure after panic recover
	lowGasCtx := infCtx.WithGasMeter(sdk.NewGasMeter(keeper.GasReserve - 1))
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(lowGasCtx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		cachedCtx.GasMeter().ConsumeGas(1, "Sudo response consumption")
	}).Return(nil, nil)
	// feeKeeper.EXPECT().DistributeAcknowledgementFee(lowGasCtx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().AddContractFailure(lowGasCtx, "channel-0", contractAddress.String(), p.GetSequence(), "ack").Do(func(ctx sdk.Context, channelId string, address string, ackID uint64, ackType string) {
		ctx.GasMeter().ConsumeGas(keeper.GasReserve, "out of gas")
	})
	require.Panics(t, func() { icak.HandleAcknowledgement(lowGasCtx, p, resAckData, relayerAddress) })
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	icak, infCtx := testkeeper.InterchainTxsKeeper(t, cmKeeper, feeKeeper, icaKeeper, nil, nil)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.PortPrefix + testutil.TestOwnerAddress + ".ica0",
		SourceChannel: "channel-0",
	}

	err := icak.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	require.ErrorContains(t, err, "failed to get ica owner from port")

	gasReserved := false
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+keeper.GasReserve {
			gasReserved = true
		}
		cachedCtx.GasMeter().ConsumeGas(1_000_000, "Sudo timeout consumption")
	}).Return(nil, nil)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	consumedBefore := ctx.GasMeter().GasConsumed()
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.True(t, gasReserved)
	require.Equal(t, consumedBefore+1_000_000, ctx.GasMeter().GasConsumed())
	require.NoError(t, err)

	// error during SudoTimeOut
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "timeout")
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// out of gas during SudoTimeOut
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		ctx.GasMeter().ConsumeGas(ctx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "timeout")
	// feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
}

func TestHandleChanOpenAck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, cmKeeper, nil, nil, nil, nil)
	portID := icatypes.PortPrefix + testutil.TestOwnerAddress + ".ica0"
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	channelID := "channel-0"
	counterpartyChannelID := "channel-1"

	err := icak.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	require.ErrorContains(t, err, "failed to get ica owner from port")

	cmKeeper.EXPECT().SudoOnChanOpenAck(ctx, contractAddress, types.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   "1",
	}).Return(nil, fmt.Errorf("SudoOnChanOpenAck error"))
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.ErrorContains(t, err, "failed to Sudo the contract OnChanOpenAck")

	cmKeeper.EXPECT().SudoOnChanOpenAck(ctx, contractAddress, types.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   "1",
	}).Return(nil, nil)
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.NoError(t, err)
}
