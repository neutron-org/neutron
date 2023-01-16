package keeper_test

import (
	"testing"

	testkeeper "github.com/neutron-org/neutron/testutil/feerefunder/keeper"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/x/feerefunder/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FeeKeeper(t, nil, nil)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
