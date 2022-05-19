package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/lidofinance/gaia-wasm-zone/testutil/interchainqueries/keeper"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/keeper"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InterchainQueriesKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
