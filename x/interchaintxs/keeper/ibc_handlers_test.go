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
	"github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

var (
	ShouldNotBeWrittenKey = []byte("shouldnotkey")
	ShouldNotBeWritten    = []byte("should not be written")
	ShouldBeWritten       = []byte("should be written")
)

func ShouldBeWrittenKey(suffix string) []byte {
	return append([]byte("shouldkey"), []byte(suffix)...)
}

func TestHandleAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	icak, infCtx, storeKey := testkeeper.InterchainTxsKeeper(t, cmKeeper, feeKeeper, icaKeeper, nil)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	store := ctx.KVStore(storeKey)

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

	// error during SudoResponse
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten) // consumes 2990
	}).Return(nil, fmt.Errorf("SudoResponse error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, p, contractAddress.String(), "ack", resACK.GetResult(), resACK.GetError())
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(2990), ctx.GasMeter().GasConsumed())

	// error during SudoError
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, err string) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, p, contractAddress.String(), "ack", errACK.GetResult(), errACK.GetError())
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(2990), ctx.GasMeter().GasConsumed())

	//  success during SudoError
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, err string) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldBeWrittenKey("sudoerror"), ShouldBeWritten)
	}).Return(nil, nil)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudoerror")))

	// out of gas during SudoError
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, error string) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(cachedCtx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, p, contractAddress.String(), "ack", errACK.GetResult(), errACK.GetError())
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed()) // due to out of gas recovery we consume 0 with a SudoError handler

	// check we have ReserveGas reserved and
	// check gas consumption from cachedCtx has added to the main ctx
	// one of the ways to check it - make the check during SudoResponse call
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	gasReserved := false
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+keeper.GasReserve {
			gasReserved = true
		}
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldBeWrittenKey("sudoresponse"), ShouldBeWritten) // consumes 3140 gas, 2000 flat write + 30 every byte of key+value
	}).Return(nil, nil)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.True(t, gasReserved)
	require.Equal(t, uint64(3140), ctx.GasMeter().GasConsumed())
	require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudoresponse")))

	// not enough gas to reserve + not enough to make AddContractFailure failure after panic recover
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	lowGasCtx := infCtx.WithGasMeter(sdk.NewGasMeter(keeper.GasReserve - 1))
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(lowGasCtx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(1, "Sudo response consumption")
	}).Return(nil, nil)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(lowGasCtx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	cmKeeper.EXPECT().AddContractFailure(lowGasCtx, p, contractAddress.String(), "ack", resACK.GetResult(), resACK.GetError()).Do(func(ctx sdk.Context, packet channeltypes.Packet, address string, ackType string, ackResult []byte, errorText string) {
		ctx.GasMeter().ConsumeGas(keeper.GasReserve, "out of gas")
	})
	require.Panics(t, func() { icak.HandleAcknowledgement(lowGasCtx, p, resAckData, relayerAddress) }) //nolint:errcheck // this is a panic test
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	icak, infCtx, storeKey := testkeeper.InterchainTxsKeeper(t, cmKeeper, feeKeeper, icaKeeper, nil)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	store := ctx.KVStore(storeKey)
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

	gasReserved := false
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+keeper.GasReserve {
			gasReserved = true
		}
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldBeWrittenKey("sudotimeout"), ShouldBeWritten) // consumes 3110 gas, 2000 flat write + 30 every byte of key+value
	}).Return(nil, nil)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.True(t, gasReserved)
	require.Equal(t, uint64(3110), ctx.GasMeter().GasConsumed())
	require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudotimeout")))
	require.NoError(t, err)

	// error during SudoTimeOut
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, p, contractAddress.String(), "timeout", []byte{}, "")
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))

	// out of gas during SudoTimeOut
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(cachedCtx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, p, contractAddress.String(), "timeout", []byte{}, "")
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
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
