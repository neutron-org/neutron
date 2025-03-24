package keeper_test

import (
	"testing"

	"github.com/neutron-org/neutron/v6/app/config"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v6/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

func TestGetParams(t *testing.T) {
	_ = config.GetDefaultConfig()

	k, ctx := testkeeper.FeeburnerKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}
