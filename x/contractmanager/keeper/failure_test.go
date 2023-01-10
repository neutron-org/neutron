package keeper_test

import (
	"crypto/rand"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/testutil/contractmanager/nullify"
	"github.com/neutron-org/neutron/x/contractmanager/keeper"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNFailure(keeper *keeper.Keeper, ctx sdk.Context, addresses int, failures int) [][]types.Failure {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	items := make([][]types.Failure, addresses)
	for i := range items {
		items[i] = make([]types.Failure, failures)
		rand.Read(pub.Key)
		acc := sdktypes.AccAddress(pub.Address())

		for c := range items[i] {
			items[i][c].Address = acc.String()
			items[i][c].Id = uint64(c)

			keeper.AddContractFailure(ctx, items[i][c].Address, 0, "")
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
	items := createNFailure(keeper, ctx, 10, 4)
	flattenItems := flattenFailures(items)

	allFailures := keeper.GetAllFailures(ctx)

	require.ElementsMatch(t,
		nullify.Fill(flattenItems),
		nullify.Fill(allFailures),
	)
}
