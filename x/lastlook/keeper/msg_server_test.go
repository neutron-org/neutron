package keeper_test

import (
	"testing"

	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite KeeperTestSuite) TestMsgUpdateParams() { //nolint:govet // it's a test so it's okay to copy locks
	msgSrv := suite.GetNeutronZoneApp(suite.ChainA).LastLookKeeper

	// update params from non-authority
	ctx := suite.ChainA.GetContext()
	resp, err := msgSrv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw",
		Params:    types.DefaultParams(),
	})
	suite.Nil(resp)
	suite.ErrorContains(err, "invalid authority")

	newParams := types.DefaultParams()

	// everything is ok
	_, err = msgSrv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Params:    newParams,
	})
	suite.NoError(err)
	params := msgSrv.GetParams(ctx)
	suite.Equal(newParams, params)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
