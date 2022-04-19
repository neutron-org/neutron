package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/lidofinance/interchain-adapter/testutil/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchainadapter/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchainadapter/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InterchainadapterKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
