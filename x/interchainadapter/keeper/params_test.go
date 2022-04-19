package keeper_test

import (
	"testing"

	testkeeper "github.com/lidofinance/interchain-adapter/testutil/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchainadapter/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.InterchainadapterKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
