package keeper_test

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/neutron-org/neutron/testutil"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/contractmanager/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/neutron-org/neutron/testutil/contractmanager/nullify"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/x/contractmanager/keeper"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNFailure(keeper *keeper.Keeper, ctx sdk.Context, addresses, failures int) [][]types.Failure {
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
			items[i][c].Id = uint64(c)
			items[i][c].Packet = &p
			items[i][c].Ack = nil
			keeper.AddContractFailure(ctx, &p, items[i][c].Address, "", nil)
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
	items := createNFailure(k, ctx, 1, 1)
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
	k.AddContractFailure(ctx, &channeltypes.Packet{}, contractAddress.String(), "ack", &channeltypes.Acknowledgement{})
	failure, err := k.GetFailure(ctx, contractAddress, failureID)
	require.NoError(t, err)
	require.Equal(t, failureID, failure.Id)
	require.Equal(t, "ack", failure.AckType)

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
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "ack", &ack)

	// success response
	xSuc := types.MessageResponse{}
	xSuc.Response.Data = data
	xSuc.Response.Request = channeltypes.Packet{}
	msgSuc, err := json.Marshal(xSuc)
	require.NoError(t, err)
	// error response
	xErr := types.MessageError{}
	xErr.Error.Request = channeltypes.Packet{}
	xErr.Error.Details = "not able to do IBC tx"
	msgErr, err := json.Marshal(xErr)
	require.NoError(t, err)
	// timeout response
	xTimeout := types.MessageTimeout{}
	xTimeout.Timeout.Request = channeltypes.Packet{}
	msgTimeout, err := json.Marshal(xTimeout)
	require.NoError(t, err)

	// case: successful resubmit with ack and ack = response
	wk.EXPECT().HasContractInfo(gomock.AssignableToTypeOf(ctx), contractAddr).Return(true)
	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgSuc).Return([]byte{}, nil)

	failure, err := k.GetFailure(ctx, contractAddr, failureID)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure)
	require.NoError(t, err)
	// failure should be deleted
	_, err = k.GetFailure(ctx, contractAddr, failureID)
	require.ErrorContains(t, err, "key not found")

	// case: failed resubmit with ack and ack = response
	failureID2 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "ack", &ack)

	wk.EXPECT().HasContractInfo(gomock.AssignableToTypeOf(ctx), contractAddr).Return(true)
	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgSuc).Return(nil, fmt.Errorf("failed to Sudo"))

	failure2, err := k.GetFailure(ctx, contractAddr, failureID2)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure2)
	require.ErrorContains(t, err, "cannot resubmit failure ack response")
	// failure is still there
	failureAfter2, err := k.GetFailure(ctx, contractAddr, failureID2)
	require.NoError(t, err)
	require.Equal(t, failureAfter2.Id, failure2.Id)

	// case: successful resubmit with ack and ack = error
	// add error failure
	failureID3 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "ack", &ackError)

	wk.EXPECT().HasContractInfo(gomock.AssignableToTypeOf(ctx), contractAddr).Return(true)
	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgErr).Return([]byte{}, nil)

	failure3, err := k.GetFailure(ctx, contractAddr, failureID3)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure3)
	require.NoError(t, err)
	// failure should be deleted
	_, err = k.GetFailure(ctx, contractAddr, failureID3)
	require.ErrorContains(t, err, "key not found")

	// case: failed resubmit with ack and ack = error
	failureID4 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "ack", &ackError)

	wk.EXPECT().HasContractInfo(gomock.AssignableToTypeOf(ctx), contractAddr).Return(true)
	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgErr).Return(nil, fmt.Errorf("failed to Sudo"))

	failure4, err := k.GetFailure(ctx, contractAddr, failureID4)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure4)
	require.ErrorContains(t, err, "cannot resubmit failure ack error")
	// failure is still there
	failureAfter4, err := k.GetFailure(ctx, contractAddr, failureID4)
	require.NoError(t, err)
	require.Equal(t, failureAfter4.Id, failure4.Id)

	// case: successful resubmit with timeout
	// add error failure
	failureID5 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "timeout", nil)

	wk.EXPECT().HasContractInfo(gomock.AssignableToTypeOf(ctx), contractAddr).Return(true)
	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgTimeout).Return([]byte{}, nil)

	failure5, err := k.GetFailure(ctx, contractAddr, failureID5)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure5)
	require.NoError(t, err)
	// failure should be deleted
	_, err = k.GetFailure(ctx, contractAddr, failureID5)
	require.ErrorContains(t, err, "key not found")

	// case: failed resubmit with timeout
	failureID6 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "timeout", nil)

	wk.EXPECT().HasContractInfo(gomock.AssignableToTypeOf(ctx), contractAddr).Return(true)
	wk.EXPECT().Sudo(gomock.AssignableToTypeOf(ctx), contractAddr, msgTimeout).Return(nil, fmt.Errorf("failed to Sudo"))

	failure6, err := k.GetFailure(ctx, contractAddr, failureID6)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure6)
	require.ErrorContains(t, err, "cannot resubmit failure ack timeout")
	// failure is still there
	failureAfter6, err := k.GetFailure(ctx, contractAddr, failureID6)
	require.NoError(t, err)
	require.Equal(t, failureAfter6.Id, failure6.Id)

	// no Failure.Ack field found for ackType = 'ack'
	failureID7 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, &packet, contractAddr.String(), "ack", nil)

	failure7, err := k.GetFailure(ctx, contractAddr, failureID7)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure7)
	require.ErrorContains(t, err, "cannot resubmit failure without acknowledgement")

	// no Failure.Packet found
	failureID8 := k.GetNextFailureIDKey(ctx, contractAddr.String())
	k.AddContractFailure(ctx, nil, contractAddr.String(), "ack", nil)

	failure8, err := k.GetFailure(ctx, contractAddr, failureID8)
	require.NoError(t, err)
	err = k.ResubmitFailure(ctx, contractAddr, failure8)
	require.ErrorContains(t, err, "cannot resubmit failure without packet info")
}
