package transfer_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/neutron-org/neutron/testutil"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/transfer/types"
	testkeeper "github.com/neutron-org/neutron/testutil/transfer/keeper"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchaintxs/keeper"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/neutron-org/neutron/x/transfer"
	"github.com/stretchr/testify/require"
)

const TestCosmosAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

func TestHandleAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	chanKeeper := mock_types.NewMockChannelKeeper(ctrl)
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return([]byte("address"))
	txKeeper, infCtx := testkeeper.TransferKeeper(t, cmKeeper, feeKeeper, chanKeeper, authKeeper)
	txModule := transfer.NewIBCModule(*txKeeper)
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
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)

	err = txModule.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	require.ErrorContains(t, err, "cannot unmarshal ICS-20 transfer packet acknowledgement")

	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.ErrorContains(t, err, "cannot unmarshal ICS-20 transfer packet data")

	token := transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   "nonbech32",
		Receiver: TestCosmosAddress,
	}
	tokenBz, err := ictxtypes.ModuleCdc.MarshalJSON(&token)
	p.Data = tokenBz

	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.ErrorContains(t, err, "failed to decode address from bech32")

	token = transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   testutil.TestOwnerAddress,
		Receiver: TestCosmosAddress,
	}
	tokenBz, err = ictxtypes.ModuleCdc.MarshalJSON(&token)
	p.Data = tokenBz

	// error during SudoResponse non contract
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Return(nil, fmt.Errorf("SudoResponse error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// error during SudoResponse contract
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Return(nil, fmt.Errorf("SudoResponse error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// error during SudoError non contract
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	// feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// error during SudoError contract
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// success during SudoError non contract
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Return(nil, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// success during SudoError contract
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Return(nil, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// recoverable out of gas during SudoError non contract
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, error string) {
		ctx.GasMeter().ConsumeGas(ctx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	// FIXME: fix distribution during outofgas
	// cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// recoverable out of gas during SudoError contract
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, error string) {
		ctx.GasMeter().ConsumeGas(ctx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "ack")
	// FIXME: fix distribution during outofgas
	// cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	// feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)

	// check we have ReserveGas reserved and
	// check gas consumption from cachedCtx has added to the main ctx
	// one of the ways to check it - make the check during SudoResponse call
	// non contract
	gasReserved := false
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+transfer.GasReserve {
			gasReserved = true
		}
		cachedCtx.GasMeter().ConsumeGas(1_000_000, "Sudo response consumption")
	}).Return(nil, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	// feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	consumedBefore := ctx.GasMeter().GasConsumed()
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.True(t, gasReserved)
	require.Equal(t, consumedBefore+1_000_000, ctx.GasMeter().GasConsumed())

	// contract
	// refresh ctx
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	gasReserved = false
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+transfer.GasReserve {
			gasReserved = true
		}
		cachedCtx.GasMeter().ConsumeGas(1_000_000, "Sudo response consumption")
	}).Return(nil, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	consumedBefore = ctx.GasMeter().GasConsumed()
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
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
	require.Panics(t, func() { txModule.HandleAcknowledgement(lowGasCtx, p, resAckData, relayerAddress) })
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	chanKeeper := mock_types.NewMockChannelKeeper(ctrl)
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return([]byte("address"))
	txKeeper, infCtx := testkeeper.TransferKeeper(t, cmKeeper, feeKeeper, chanKeeper, authKeeper)
	txModule := transfer.NewIBCModule(*txKeeper)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	err := txModule.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	require.ErrorContains(t, err, "cannot unmarshal ICS-20 transfer packet data")

	token := transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   "nonbech32",
		Receiver: TestCosmosAddress,
	}
	tokenBz, err := ictxtypes.ModuleCdc.MarshalJSON(&token)
	p.Data = tokenBz
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.ErrorContains(t, err, "failed to decode address from bech32")

	// success non contract
	token = transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   testutil.TestOwnerAddress,
		Receiver: TestCosmosAddress,
	}
	tokenBz, err = ictxtypes.ModuleCdc.MarshalJSON(&token)
	p.Data = tokenBz
	gasReserved := false
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+keeper.GasReserve {
			gasReserved = true
		}
		cachedCtx.GasMeter().ConsumeGas(1_000_000, "Sudo timeout consumption")
	}).Return(nil, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	consumedBefore := ctx.GasMeter().GasConsumed()
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.True(t, gasReserved)
	require.Equal(t, consumedBefore+1_000_000, ctx.GasMeter().GasConsumed())
	require.NoError(t, err)

	// success contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	gasReserved = false
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		if ctx.GasMeter().Limit() == cachedCtx.GasMeter().Limit()+keeper.GasReserve {
			gasReserved = true
		}
		cachedCtx.GasMeter().ConsumeGas(1_000_000, "Sudo timeout consumption")
	}).Return(nil, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	consumedBefore = ctx.GasMeter().GasConsumed()
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.True(t, gasReserved)
	require.Equal(t, consumedBefore+1_000_000, ctx.GasMeter().GasConsumed())
	require.NoError(t, err)

	// error during SudoTimeOut non contract
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "timeout")
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// error during SudoTimeOut contract
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "timeout")
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// out of gas during SudoTimeOut non contract
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		ctx.GasMeter().ConsumeGas(ctx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "timeout")
	// cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// out of gas during SudoTimeOut contract
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(ctx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		ctx.GasMeter().ConsumeGas(ctx.GasMeter().Limit()+1, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, "channel-0", contractAddress.String(), p.GetSequence(), "timeout")
	// FIXME: make DistributeTimeoutFee during out of gas
	// cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	// feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
}
