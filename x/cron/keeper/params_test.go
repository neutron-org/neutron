package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	testkeeper "github.com/neutron-org/neutron/testutil/cron/keeper"
	"github.com/neutron-org/neutron/x/cron/types"
)

func TestGetParams(t *testing.T) {
	_ = app.GetDefaultConfig()

	k, ctx := testkeeper.CronKeeper(t, nil, nil)
	params := types.Params{
		SecurityAddress: testutil.TestOwnerAddress,
		Limit:           5,
	}

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
