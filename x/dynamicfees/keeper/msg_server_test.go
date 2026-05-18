package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v11/testutil"
	"github.com/neutron-org/neutron/v11/x/dynamicfees/types"
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite KeeperTestSuite) TestMsgUpdateParams() { //nolint:govet // it's a test so it's okay to copy locks
	msgSrv := suite.GetNeutronZoneApp(suite.ChainA).DynamicFeesKeeper

	// update params from non-authority
	ctx := suite.ChainA.GetContext()
	resp, err := msgSrv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw",
		Params:    types.DefaultParams(),
	})
	suite.Nil(resp)
	suite.ErrorContains(err, "invalid authority")

	newParams := types.DefaultParams()
	newParams.NtrnPrices = append(newParams.NtrnPrices, cosmostypes.DecCoin{Denom: "uatom", Amount: math.LegacyMustNewDecFromStr("0.1")})

	// everything is ok
	_, err = msgSrv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAddress(govtypes.ModuleName).String(),
		Params:    newParams,
	})
	suite.NoError(err)
	params := msgSrv.GetParams(ctx)
	suite.Equal(newParams, params)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
