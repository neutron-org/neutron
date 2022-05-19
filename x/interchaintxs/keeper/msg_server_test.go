package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/lidofinance/gaia-wasm-zone/testutil/interchaintxs/keeper"
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/keeper"
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InterchainTxsKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
