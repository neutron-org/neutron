package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/testutil/interchainqueries/keeper"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
