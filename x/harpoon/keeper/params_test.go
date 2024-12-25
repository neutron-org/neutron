package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

    keepertest "github.com/neutron-org/neutron/v5/testutil/keeper"
    "github.com/neutron-org/neutron/v5/x/harpoon/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.HarpoonKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
