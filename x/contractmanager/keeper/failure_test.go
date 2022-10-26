package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/neutron-org/neutron/testutil/contractmanager/keeper"
	"github.com/neutron-org/neutron/testutil/contractmanager/nullify"
	"github.com/neutron-org/neutron/x/contractmanager/keeper"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNFailure(keeper *keeper.Keeper, ctx sdk.Context, addresses int, failures int) [][]types.Failure {
	items := make([][]types.Failure, addresses)
	for i := range items {
		items[i] = make([]types.Failure, failures)
		for c := range items[i] {
			items[i][c].Address = fmt.Sprintf("address%d", i)
			items[i][c].Id = uint64(c)

			keeper.AddContractFailure(ctx, items[i][c])
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

func TestFailureGet(t *testing.T) {
	keeper, ctx := keepertest.ContractmanagerKeeper(t)
	items := createNFailure(keeper, ctx, 10, 4)
	for _, item := range items {
		rst := keeper.GetContractFailures(ctx,
			item[0].Address,
		)
		require.Equal(t,
			nullify.Fill(item),
			nullify.Fill(&rst),
		)
	}
}

func TestFailureGetAll(t *testing.T) {
	keeper, ctx := keepertest.ContractmanagerKeeper(t)
	items := createNFailure(keeper, ctx, 10, 4)
	flattenItems := flattenFailures(items)

	allFailures := keeper.GetAllFailures(ctx)

	require.ElementsMatch(t,
		nullify.Fill(flattenItems),
		nullify.Fill(allFailures),
	)
}
