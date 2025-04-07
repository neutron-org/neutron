package keeper_test

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"

	"github.com/golang/mock/gomock"

	"github.com/neutron-org/neutron/v6/testutil"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/contractmanager/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v6/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/v6/x/contractmanager/keeper"
	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNFailure(k *keeper.Keeper, ctx sdk.Context, addresses, failures int) [][]types.Failure {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	items := make([][]types.Failure, addresses)
	for i := range items {
		items[i] = make([]types.Failure, failures)
		rand.Read(pub.Key) //nolint:errcheck
		acc := sdk.AccAddress(pub.Address())

		for c := range items[i] {
			p := channeltypes.Packet{
				Sequence:   0,
				SourcePort: "port-n",
			}
			items[i][c].Address = acc.String()
			items[i][c].Id = uint64(c) //nolint:gosec
			sudo, err := keeper.PrepareSudoCallbackMessage(p, nil)
			if err != nil {
				panic(err)
			}
			items[i][c].SudoPayload = sudo
			items[i][c].Error = "test error"
			k.AddContractFailure(ctx, items[i][c].Address, sudo, "test error")
		}
	}
	return items
}

func flattenFailures(items [][]types.Failure) []types.Failure {
	m := len(items)
	n := len(items[0])

	flattenItems := make([]types.Failure, m*n)
	for i, failures := range items {
		for c, failure := range failures {
			flattenItems[i*n+c] = failure
		}
	}

	return flattenItems
}

func TestGetAllFailures(t *testing.T) {
	k, ctx := keepertest.ContractManagerKeeper(t, nil)
	items := createNFailure(k, ctx, 10, 4)
	flattenItems := flattenFailures(items)

	allFailures := k.GetAllFailures(ctx)

	require.ElementsMatch(t,
		nullify.Fill(flattenItems),
		nullify.Fill(allFailures),
	)
}

func TestAddGetFailure(t *testing.T) {
	// test adding and getting failure
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	k, ctx := keepertest.ContractManagerKeeper(t, nil)
	failureID := k.GetNextFailureIDKey(ctx, contractAddress.String())
	sudoPayload := []byte("payload")
	k.AddContractFailure(ctx, contractAddress.String(), sudoPayload, "test error")
	failure, err := k.GetFailure(ctx, contractAddress, failureID)
	require.NoError(t, err)
	require.Equal(t, failureID, failure.Id)
	require.Equal(t, sudoPayload, failure.SudoPayload)
	require.Equal(t, "test error", failure.Error)

	// non-existent id
	_, err = k.GetFailure(ctx, contractAddress, failureID+1)
	require.Error(t, err)

	// non-existent contract address
	_, err = k.GetFailure(ctx, sdk.MustAccAddressFromBech32("neutron1nseacn2aqezhj3ssatfg778ctcfjuknm8ucc0l"), failureID)
	require.Error(t, err)
}

func TestResubmitFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	wk := mock_types.NewMockWasmKeeper(ctrl)
	k, ctx := keepertest.ContractManagerKeeper(t, wk)

	contractAddr := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	data := []byte("Result")
	ack := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{Result: data},
	}
	ackError := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{Error: "not able to do IBC tx"},
	}

	// add ack failure
	packet := channeltypes.Packet{}
	failureID := k.GetNextFailureIDKey(ctx, contractAddr.String())
	payload, err := keeper.PrepareSudoCallbackMessage(packet, &ack)
	require.NoError(t, err)
	k.AddContractFailure(ctx, contractAddr.String(), payload, "test error")

	// success response
	xSuc := types.MessageSudoCallback{Response: &types.ResponseSudoPayload{
		Request: channeltypes.Packet{},
		Data:    data,
	}}
	msgSuc, err := json.Marshal(xSuc)
	require.NoError(t, err)
	// error response
	xErr := types.MessageSudoCallback{Error: &types.ErrorSudoPayload{
		Request: channeltypes.Packet{},
		Details: "not able to do IBC tx",
	}}
	msgErr, err := json.Marshal(xErr)
	require.NoError(t, err)
	// timeout response
	xTimeout := types.MessageSudoCallback{Timeout: &types.TimeoutPayload{Request: channeltypes.Packet{}}}
	msgTimeout, err := json.Marshal(xTimeout)
	require.NoError(t, err)

	// case: successful resubmit with ack and ack = response
	wk.EXPECT().Sudo(ctx, contractAddr, msgSuc).Return([]byte{}, nil)
	wk.EXPECT().HasContractInfo(ctx, contractAddr).Return(true)

	failure, err := k.GetFailure(ctx, contractAddr, failureID)
	require.NoError(t, err)
	_, err = k.ResubmitFailure(ctx, &types.MsgResubmitFailure{
		Sender:    contractAddr.String(),
		FailureId: failure.Id,
	})
	require.NoError(t, err)
	// failure should be deleted
	_, err = k.GetFailure(ctx, contractAddr, failureID)
	require.ErrorContains(t, err, "key not found")

	// case: failed resubmit with ack and ack = response
	failureID2 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	payload, err = keeper.PrepareSudoCallbackMessage(packet, &ack)
	require.NoError(t, err)
	k.AddContractFailure(ctx, contractAddr.String(), payload, "test error")

	wk.EXPECT().Sudo(ctx, contractAddr, msgSuc).Return(nil, fmt.Errorf("failed to sudo"))
	wk.EXPECT().HasContractInfo(ctx, contractAddr).Return(true)

	failure2, err := k.GetFailure(ctx, contractAddr, failureID2)
	require.NoError(t, err)
	_, err = k.ResubmitFailure(ctx, &types.MsgResubmitFailure{
		Sender:    contractAddr.String(),
		FailureId: failure2.Id,
	})
	require.ErrorContains(t, err, "cannot resubmit failure")
	// failure is still there
	failureAfter2, err := k.GetFailure(ctx, contractAddr, failureID2)
	require.NoError(t, err)
	require.Equal(t, failureAfter2.Id, failure2.Id)

	// case: successful resubmit with ack and ack = error
	// add error failure
	failureID3 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	payload, err = keeper.PrepareSudoCallbackMessage(packet, &ackError)
	require.NoError(t, err)
	k.AddContractFailure(ctx, contractAddr.String(), payload, "test error")

	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgErr).Return([]byte{}, nil)
	wk.EXPECT().HasContractInfo(ctx, contractAddr).Return(true)

	failure3, err := k.GetFailure(ctx, contractAddr, failureID3)
	require.NoError(t, err)
	_, err = k.ResubmitFailure(ctx, &types.MsgResubmitFailure{
		Sender:    contractAddr.String(),
		FailureId: failure3.Id,
	})
	require.NoError(t, err)
	// failure should be deleted
	_, err = k.GetFailure(ctx, contractAddr, failureID3)
	require.ErrorContains(t, err, "key not found")

	// case: failed resubmit with ack and ack = error
	failureID4 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	payload, err = keeper.PrepareSudoCallbackMessage(packet, &ackError)
	require.NoError(t, err)
	k.AddContractFailure(ctx, contractAddr.String(), payload, "test error")

	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgErr).Return(nil, fmt.Errorf("failed to sudo"))
	wk.EXPECT().HasContractInfo(ctx, contractAddr).Return(true)

	failure4, err := k.GetFailure(ctx, contractAddr, failureID4)
	require.NoError(t, err)
	_, err = k.ResubmitFailure(ctx, &types.MsgResubmitFailure{
		Sender:    contractAddr.String(),
		FailureId: failure4.Id,
	})
	require.ErrorContains(t, err, "cannot resubmit failure")
	// failure is still there
	failureAfter4, err := k.GetFailure(ctx, contractAddr, failureID4)
	require.NoError(t, err)
	require.Equal(t, failureAfter4.Id, failure4.Id)

	// case: successful resubmit with timeout
	// add error failure
	failureID5 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	payload, err = keeper.PrepareSudoCallbackMessage(packet, nil)
	require.NoError(t, err)
	k.AddContractFailure(ctx, contractAddr.String(), payload, "test error")

	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgTimeout).Return([]byte{}, nil)
	wk.EXPECT().HasContractInfo(ctx, contractAddr).Return(true)

	failure5, err := k.GetFailure(ctx, contractAddr, failureID5)
	require.NoError(t, err)
	_, err = k.ResubmitFailure(ctx, &types.MsgResubmitFailure{
		Sender:    contractAddr.String(),
		FailureId: failure5.Id,
	})
	require.NoError(t, err)
	// failure should be deleted
	_, err = k.GetFailure(ctx, contractAddr, failureID5)
	require.ErrorContains(t, err, "key not found")

	// case: failed resubmit with timeout
	failureID6 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	payload, err = keeper.PrepareSudoCallbackMessage(packet, nil)
	require.NoError(t, err)
	k.AddContractFailure(ctx, contractAddr.String(), payload, "test error")

	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgTimeout).Return(nil, fmt.Errorf("failed to sudo"))
	wk.EXPECT().HasContractInfo(ctx, contractAddr).Return(true)

	failure6, err := k.GetFailure(ctx, contractAddr, failureID6)
	require.NoError(t, err)
	_, err = k.ResubmitFailure(ctx, &types.MsgResubmitFailure{
		Sender:    contractAddr.String(),
		FailureId: failure6.Id,
	})
	require.ErrorContains(t, err, "cannot resubmit failure")
	// failure is still there
	failureAfter6, err := k.GetFailure(ctx, contractAddr, failureID6)
	require.NoError(t, err)
	require.Equal(t, failureAfter6.Id, failure6.Id)
}
