package keeper_test

import (
	"crypto/rand"
	"strconv"
	"testing"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/neutron-org/neutron/testutil"
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
				SourcePort:    icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ".ica0", // TODO: maybe change
				SourceChannel: items[i][c].ChannelId,
			}
			items[i][c].Address = acc.String()
			items[i][c].Id = uint64(c)
			items[i][c].Packet = &p
			items[i][c].AckResult = []byte{}
			items[i][c].ErrorText = ""

			// TODO
			keeper.AddContractFailure(ctx, p, items[i][c].Address, "", []byte{}, "")
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
	keeper, ctx := keepertest.ContractManagerKeeper(t, nil)
	items := createNFailure(keeper, ctx, 1, 1)
	flattenItems := flattenFailures(items)

	allFailures := keeper.GetAllFailures(ctx)

	require.ElementsMatch(t,
		nullify.Fill(flattenItems),
		// flattenItems,
		// allFailures,
		nullify.Fill(allFailures),
	)
}
