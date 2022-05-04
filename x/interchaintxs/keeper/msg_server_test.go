package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/lidofinance/interchain-adapter/testutil/interchaintxs/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchaintxs/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchaintxs/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InterchainQueriesKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
