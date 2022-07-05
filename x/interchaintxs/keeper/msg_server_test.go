package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	"github.com/neutron-org/neutron/x/interchaintxs/keeper"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InterchainTxsKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
