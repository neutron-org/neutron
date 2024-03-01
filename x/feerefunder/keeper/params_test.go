package keeper_test

import (
	"testing"

	testkeeper "github.com/neutron-org/neutron/v3/testutil/feerefunder/keeper"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v3/x/feerefunder/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FeeKeeper(t, nil, nil)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}

	require.EqualValues(t, params, k.GetParams(ctx))
}
