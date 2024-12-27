package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

func setupMsgServer(t testing.TB) (keeper.Keeper, types.MsgServer, context.Context) {
	k, ctx := keepertest.HarpoonKeeper(t)
	return *k, keeper.NewMsgServerImpl(*k), ctx
}

func TestMsgServer(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
	require.NotEmpty(t, k)
}
