package keeper_test

import (
	"crypto/rand"
	"github.com/neutron-org/neutron/testutil"
	"strconv"
	"testing"

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
				Sequence:      0,
				SourcePort:    "port-n",
				SourceChannel: items[i][c].ChannelId,
			}
			items[i][c].Address = acc.String()
			items[i][c].Id = uint64(c)
			items[i][c].Packet = &p
			items[i][c].Ack = nil
			keeper.AddContractFailure(ctx, p, items[i][c].Address, "", nil)
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
	failureId := k.GetNextFailureIDKey(ctx, contractAddress.String())
	k.AddContractFailure(ctx, channeltypes.Packet{}, contractAddress.String(), "ack", &channeltypes.Acknowledgement{})
	failure, err := k.GetFailure(ctx, contractAddress, failureId)
	require.NoError(t, err)
	require.Equal(t, failureId, failure.Id)
	require.Equal(t, "ack", failure.AckType)

	// non-existent id
	_, err = k.GetFailure(ctx, contractAddress, failureId+1)
	require.Error(t, err)

	// non-existent contract address
	_, err = k.GetFailure(ctx, sdk.MustAccAddressFromBech32("neutron1nseacn2aqezhj3ssatfg778ctcfjuknm8ucc0l"), failureId)
	require.Error(t, err)
}

func TestResubmitFailure(t *testing.T) {
	// TODO

	// successful resubmit with ack and ack = response
	// failed resubmit with ack and ack = response

	// successful resubmit with ack and ack = error
	// failed resubmit with ack and ack = error

	// successful resubmit with timeout
	// failed resubmit with timeout

	// no Failure.Ack field found for ackType = 'ack'
	// no Failure.Packet found
}
