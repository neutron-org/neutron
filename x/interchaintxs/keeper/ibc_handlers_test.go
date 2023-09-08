package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/testutil"
	testkeeper "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/interchaintxs/types"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
)

func TestHandleAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	icak, infCtx, _ := testkeeper.InterchainTxsKeeper(t, cmKeeper, feeKeeper, icaKeeper, nil)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	//store := ctx.KVStore(storeKey)

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
		SourcePort:    icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ".ica0",
		SourceChannel: "channel-0",
	}
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)

	err = icak.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	require.ErrorContains(t, err, "failed to get ica owner from port")

	err = icak.HandleAcknowledgement(ctx, p, nil, relayerAddress)
	require.ErrorContains(t, err, "cannot unmarshal ICS-27 packet acknowledgement")

	// success contract SudoResponse
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().SudoResponse(ctx, contractAddress, p, resACK)
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// success contract SudoError
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().SudoError(ctx, contractAddress, p, errACK)
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// error contract SudoResponse
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().SudoResponse(ctx, contractAddress, p, resACK).Return(nil, fmt.Errorf("error sudoResponse"))
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.Error(t, err)

	// error contract SudoError
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().SudoError(ctx, contractAddress, p, errACK).Return(nil, fmt.Errorf("error sudoError"))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.Error(t, err)

	// no contract SudoError
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	////  success during SudoError
	//ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	//cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, err string) {
	//	store := cachedCtx.KVStore(storeKey)
	//	store.Set(ShouldBeWrittenKey("sudoerror"), ShouldBeWritten)
	//}).Return(nil, nil)
	//cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 6000})
	//feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	//err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	//require.NoError(t, err)
	//require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudoerror")))
	//require.Equal(t, uint64(3050), ctx.GasMeter().GasConsumed())
	//
	//// out of gas during SudoError
	//ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	//cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, error string) {
	//	store := cachedCtx.KVStore(storeKey)
	//	store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
	//	cachedCtx.GasMeter().ConsumeGas(7001, "out of gas test")
	//})
	//cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 7000})
	//cmKeeper.EXPECT().AddContractFailure(ctx, &p, contractAddress.String(), types.Ack, &errACK)
	//feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	//err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	//require.NoError(t, err)
	//require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	//require.Equal(t, uint64(7000), ctx.GasMeter().GasConsumed())
	//
	//// success during SudoResponse
	//ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	//cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
	//	store := cachedCtx.KVStore(storeKey)
	//	store.Set(ShouldBeWrittenKey("sudoresponse"), ShouldBeWritten) // consumes 3140 gas, 2000 flat write + 30 every byte of key+value
	//}).Return(nil, nil)
	//cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 8000})
	//feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	//err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	//require.NoError(t, err)
	//require.Equal(t, uint64(3140), ctx.GasMeter().GasConsumed())
	//require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudoresponse")))
	//
	//// not enough gas provided by relayer for SudoCallGasLimit
	//ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	//lowGasCtx := infCtx.WithGasMeter(sdk.NewGasMeter(1000))
	//cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(lowGasCtx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
	//	store := cachedCtx.KVStore(storeKey)
	//	store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
	//	cachedCtx.GasMeter().ConsumeGas(1001, "out of gas test")
	//})
	//feeKeeper.EXPECT().DistributeAcknowledgementFee(lowGasCtx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	//cmKeeper.EXPECT().GetParams(lowGasCtx).Return(types.Params{SudoCallGasLimit: 9000})
	//require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "consume gas from cached context"}, func() { icak.HandleAcknowledgement(lowGasCtx, p, resAckData, relayerAddress) }) //nolint:errcheck // this is a panic test
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	icak, infCtx, _ := testkeeper.InterchainTxsKeeper(t, cmKeeper, feeKeeper, icaKeeper, nil)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ".ica0",
		SourceChannel: "channel-0",
	}

	err := icak.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	require.ErrorContains(t, err, "failed to get ica owner from port")

	// contract success
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().SudoTimeout(ctx, contractAddress, p)
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// contract error
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().SudoTimeout(ctx, contractAddress, p).Return(nil, fmt.Errorf("SudoTimeout error"))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.Error(t, err)

	// no contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
}

func TestHandleChanOpenAck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	icak, ctx, _ := testkeeper.InterchainTxsKeeper(t, cmKeeper, nil, nil, nil)
	portID := icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ".ica0"
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	channelID := "channel-0"
	counterpartyChannelID := "channel-1"

	err := icak.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	require.ErrorContains(t, err, "failed to get ica owner from port")

	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	cmKeeper.EXPECT().SudoOnChanOpenAck(ctx, contractAddress, types.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   "1",
	}).Return(nil, fmt.Errorf("SudoOnChanOpenAck error"))
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.ErrorContains(t, err, "failed to sudo the contract OnChanOpenAck")

	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	cmKeeper.EXPECT().SudoOnChanOpenAck(ctx, contractAddress, types.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   "1",
	}).Return(nil, nil)
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.NoError(t, err)

	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.NoError(t, err)
}
