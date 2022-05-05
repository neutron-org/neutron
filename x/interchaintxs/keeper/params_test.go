package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/lidofinance/interchain-adapter/testutil/interchaintxs/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchaintxs/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
