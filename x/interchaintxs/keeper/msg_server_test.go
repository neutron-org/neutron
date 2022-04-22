package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/lidofinance/interchain-adapter/testutil/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/keeper"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InterchainQueriesKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
