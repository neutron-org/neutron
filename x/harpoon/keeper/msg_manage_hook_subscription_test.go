package keeper_test

import (
	"github.com/golang/mock/gomock"
	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	//"context"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/harpoon/types"
	"testing"
	//"github.com/stretchr/testify/require"
	//
	//keepertest "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	//"github.com/neutron-org/neutron/v5/x/harpoon/keeper"
	//"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

func TestManageHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)

	wasmMsgServer := mock_types.NewMockWasmMsgServer(ctrl)
	k, ctx := testutil_keeper.HarpoonKeeper(t, wasmMsgServer, accountKeeper)
}
