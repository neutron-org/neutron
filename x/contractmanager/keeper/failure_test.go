package keeper_test

import (
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

func createNFailure(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Failure {
	items := make([]types.Failure, n)
	for i := range items {
		items[i].Address = strconv.Itoa(i)

		keeper.AddContractFailure(ctx, items[i])
	}
	return items
}

func TestFailureGet(t *testing.T) {
	keeper, ctx := keepertest.ContractmanagerKeeper(t)
	items := createNFailure(keeper, ctx, 10)
	for _, item := range items {
		rst := keeper.GetContractFailures(ctx,
			item.Address,
		)
		require.Equal(t,
			nullify.Fill([]types.Failure{item}),
			nullify.Fill(&rst),
		)
	}
}

func TestFailureGetAll(t *testing.T) {
	keeper, ctx := keepertest.ContractmanagerKeeper(t)
	items := createNFailure(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllFailures(ctx)),
	)
}
