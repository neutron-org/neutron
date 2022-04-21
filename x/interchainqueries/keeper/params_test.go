package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/lidofinance/interchain-adapter/testutil/interchainqueries/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.InterchainadapterKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
