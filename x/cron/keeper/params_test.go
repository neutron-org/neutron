package keeper_test

import (
	"testing"

	"github.com/neutron-org/neutron/testutil"

	"github.com/neutron-org/neutron/app"

	testkeeper "github.com/neutron-org/neutron/testutil/cron/keeper"

	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	_ = app.GetDefaultConfig()

	k, ctx := testkeeper.CronKeeper(t, nil)
	params := types.Params{
		AdminAddress:    testutil.TestOwnerAddress,
		SecurityAddress: testutil.TestOwnerAddress,
		Limit:           5,
	}

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
