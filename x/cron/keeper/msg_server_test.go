package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/neutron-org/neutron/testutil/cron/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/cron/keeper"
	"github.com/neutron-org/neutron/x/cron/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.CronKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
