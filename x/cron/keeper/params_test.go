package keeper_test

import (
	"testing"

	testkeeper "github.com/neutron-org/neutron/testutil/cron/keeper"

	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.CronKeeper(t, nil)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
