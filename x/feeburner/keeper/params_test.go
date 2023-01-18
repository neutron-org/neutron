package keeper_test

import (
	"testing"

	"github.com/neutron-org/neutron/app"

	testkeeper "github.com/neutron-org/neutron/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/x/feeburner/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	_ = app.GetDefaultConfig()

	k, ctx := testkeeper.FeeburnerKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
