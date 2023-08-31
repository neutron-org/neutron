package transfer_test

import (
	"fmt"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/neutron-org/neutron/testutil"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/transfer/types"
	testkeeper "github.com/neutron-org/neutron/testutil/transfer/keeper"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	ictxtypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/neutron-org/neutron/x/transfer"
	"github.com/stretchr/testify/require"
)

const TestCosmosAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

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
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	chanKeeper := mock_types.NewMockChannelKeeper(ctrl)
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(transfertypes.ModuleName).Return([]byte("address"))
	txKeeper, infCtx, storeKey := testkeeper.TransferKeeper(t, cmKeeper, feeKeeper, chanKeeper, authKeeper)
	txModule := transfer.NewIBCModule(*txKeeper)
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
	require.NoError(t, err)
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
	require.NoError(t, err)
	p.Data = tokenBz

	//// error during SudoResponse non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// error during SudoResponse contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten) // consumes 2990
	}).Return(nil, fmt.Errorf("SudoResponse error"))
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 5000})
	cmKeeper.EXPECT().AddContractFailure(ctx, &p, contractAddress.String(), types.Ack, &resACK)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(2990), ctx.GasMeter().GasConsumed())

	// error during SudoError non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// error during SudoError contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg string) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten) // consumes 2990
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 7000})
	cmKeeper.EXPECT().AddContractFailure(ctx, &p, contractAddress.String(), types.Ack, &errACK)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(2990), ctx.GasMeter().GasConsumed())

	// success during SudoError non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// success during SudoError contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, err string) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldBeWrittenKey("sudoerror_contract"), ShouldBeWritten)
	}).Return(nil, nil)
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 9000})
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudoerror_contract")))
	require.Equal(t, uint64(3320), ctx.GasMeter().GasConsumed())

	// recoverable out of gas during SudoError non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	require.NoError(t, err)
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// recoverable out of gas during SudoError contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK.GetError()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, error string) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(11001, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoError error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, &p, contractAddress.String(), types.Ack, &errACK)
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 11000})
	// FIXME: fix distribution during outofgas
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, errAckData, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(11000), ctx.GasMeter().GasConsumed())

	// check we have ReserveGas reserved and
	// check gas consumption from cachedCtx has added to the main ctx
	// one of the ways to check it - make the check during SudoResponse call
	// non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	gasReserved := false
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(ctx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		if cachedCtx.GasMeter().Limit() == 13000 {
			gasReserved = true
		}
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldBeWrittenKey("sudoresponse_contract_success"), ShouldBeWritten)
	}).Return(nil, nil)
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 13000})
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
	require.True(t, gasReserved)
	require.Equal(t, uint64(3650), ctx.GasMeter().GasConsumed())
	require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudoresponse_contract_success")))

	// not enough gas provided by relayer SudoCallGasLimit
	lowGasCtx := infCtx.WithGasMeter(sdk.NewGasMeter(1000))
	cmKeeper.EXPECT().SudoResponse(gomock.AssignableToTypeOf(lowGasCtx), contractAddress, p, resACK.GetResult()).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, msg []byte) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(1001, "out of gas test")
	}).Return(nil, nil)
	cmKeeper.EXPECT().GetParams(lowGasCtx).Return(types.Params{SudoCallGasLimit: 14000})
	cmKeeper.EXPECT().HasContractInfo(lowGasCtx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(lowGasCtx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "consume gas from cached context"}, func() { txModule.HandleAcknowledgement(lowGasCtx, p, resAckData, relayerAddress) }) //nolint:errcheck // this is a test
	// NOTE: looks its impossible to test store reset after panic, because test `require.PanicsWithValue` recovers the panic
	// require.Empty(t, store.Get(ShouldNotBeWrittenKey))
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	chanKeeper := mock_types.NewMockChannelKeeper(ctrl)
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(transfertypes.ModuleName).Return([]byte("address"))
	txKeeper, infCtx, storeKey := testkeeper.TransferKeeper(t, cmKeeper, feeKeeper, chanKeeper, authKeeper)
	txModule := transfer.NewIBCModule(*txKeeper)
	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	store := ctx.KVStore(storeKey)
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
	require.NoError(t, err)
	p.Data = tokenBz
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.ErrorContains(t, err, "failed to decode address from bech32")

	token = transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   testutil.TestOwnerAddress,
		Receiver: TestCosmosAddress,
	}
	tokenBz, err = ictxtypes.ModuleCdc.MarshalJSON(&token)
	require.NoError(t, err)
	p.Data = tokenBz

	// success non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())
	require.NoError(t, err)

	// success contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	gasReserved := false
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		if cachedCtx.GasMeter().Limit() == 5000 {
			gasReserved = true
		}
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldBeWrittenKey("sudotimeout_contract_success"), ShouldBeWritten)
	}).Return(nil, nil)
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 5000})
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.True(t, gasReserved)
	require.NoError(t, err)
	require.Equal(t, uint64(3620), ctx.GasMeter().GasConsumed())
	require.Equal(t, ShouldBeWritten, store.Get(ShouldBeWrittenKey("sudotimeout_contract_success")))

	// error during SudoTimeOut non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// error during SudoTimeOut contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 7000})
	cmKeeper.EXPECT().AddContractFailure(ctx, &p, contractAddress.String(), types.Timeout, nil)
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(2990), ctx.GasMeter().GasConsumed())

	// out of gas during SudoTimeOut non contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
	require.Equal(t, uint64(0), ctx.GasMeter().GasConsumed())

	// out of gas during SudoTimeOut contract
	ctx = infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(ctx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(8001, "out of gas test")
	}).Return(nil, fmt.Errorf("SudoTimeout error"))
	cmKeeper.EXPECT().AddContractFailure(ctx, &p, contractAddress.String(), types.Timeout, nil)
	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 8000})
	cmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
	require.Empty(t, store.Get(ShouldNotBeWrittenKey))
	require.Equal(t, uint64(8000), ctx.GasMeter().GasConsumed())

	// not enough gas provided by relayer for SudoCallGasLimit
	lowGasCtx := infCtx.WithGasMeter(sdk.NewGasMeter(1000))
	cmKeeper.EXPECT().SudoTimeout(gomock.AssignableToTypeOf(lowGasCtx), contractAddress, p).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet) {
		store := cachedCtx.KVStore(storeKey)
		store.Set(ShouldNotBeWrittenKey, ShouldNotBeWritten)
		cachedCtx.GasMeter().ConsumeGas(1001, "out of gas test")
	}).Return(nil, nil)
	cmKeeper.EXPECT().GetParams(lowGasCtx).Return(types.Params{SudoCallGasLimit: 14000})
	cmKeeper.EXPECT().HasContractInfo(lowGasCtx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(lowGasCtx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "consume gas from cached context"}, func() { txModule.HandleTimeout(lowGasCtx, p, relayerAddress) }) //nolint:errcheck // this is a test
	// NOTE: looks its impossible to test store reset after panic, because test `require.PanicsWithValue` recovers the panic
	// require.Empty(t, store.Get(ShouldNotBeWrittenKey))
}
