package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/adminmodule/keeper"
	"github.com/neutron-org/neutron/x/adminmodule/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context, *keeper.Keeper) {
	k, ctx := setupKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx), k
}
