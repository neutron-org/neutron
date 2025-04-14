package transfer_test

import (
	"fmt"
	"testing"

	types2 "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v6/x/contractmanager/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/transfer/types"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/transfer/keeper"
	feetypes "github.com/neutron-org/neutron/v6/x/feerefunder/types"
	ictxtypes "github.com/neutron-org/neutron/v6/x/interchaintxs/types"
	"github.com/neutron-org/neutron/v6/x/transfer"
)

const TestCosmosAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

func TestHandleAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	chanKeeper := mock_types.NewMockChannelKeeper(ctrl)
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	tokenfactoryKeeper := mock_types.NewMockTokenfactoryKeeper(ctrl)

	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(transfertypes.ModuleName).Return([]byte("address"))
	txKeeper, infCtx, _ := testkeeper.TransferKeeper(t, wmKeeper, feeKeeper, chanKeeper, authKeeper)
	txModule := transfer.NewIBCModule(*txKeeper, wmKeeper, tokenfactoryKeeper)
	ctx := infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))

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

	msgAck, err := keeper.PrepareSudoCallbackMessage(p, &resACK)
	require.NoError(t, err)

	// non contract
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// error during Sudo contract
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msgAck).Return(nil, fmt.Errorf("SudoResponse error"))
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// success during Sudo contract
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeAcknowledgementFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msgAck)
	err = txModule.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	chanKeeper := mock_types.NewMockChannelKeeper(ctrl)
	authKeeper := mock_types.NewMockAccountKeeper(ctrl)
	tokenfactoryKeeper := mock_types.NewMockTokenfactoryKeeper(ctrl)
	// required to initialize keeper
	authKeeper.EXPECT().GetModuleAddress(transfertypes.ModuleName).Return([]byte("address"))
	txKeeper, infCtx, _ := testkeeper.TransferKeeper(t, wmKeeper, feeKeeper, chanKeeper, authKeeper)
	txModule := transfer.NewIBCModule(*txKeeper, wmKeeper, tokenfactoryKeeper)
	ctx := infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
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

	msg, err := keeper.PrepareSudoCallbackMessage(p, nil)
	require.NoError(t, err)

	// success non contract
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(false)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// success contract
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msg).Return(nil, nil)
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// error during SudoTimeOut contract
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().HasContractInfo(ctx, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
	feeKeeper.EXPECT().DistributeTimeoutFee(ctx, relayerAddress, feetypes.NewPacketID(p.SourcePort, p.SourceChannel, p.Sequence))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msg).Return(nil, fmt.Errorf("SudoTimeout error"))
	err = txModule.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
}
