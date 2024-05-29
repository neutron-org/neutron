package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v4/testutil/dynamicfees/keeper"
	"github.com/neutron-org/neutron/v4/x/dynamicfees/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.DynamicFeesKeeper(t)
	params := types.DefaultParams()

	require.EqualValues(t, params, k.GetParams(ctx))
}
