package keeper_test

import (
	"testing"

	"github.com/neutron-org/neutron/v2/app"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v2/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/v2/x/feeburner/types"
)

func TestGetParams(t *testing.T) {
	_ = app.GetDefaultConfig()

	k, ctx := testkeeper.FeeburnerKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}
