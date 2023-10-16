package cli_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/testutil/apptesting"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/keeper"
	"github.com/neutron-org/neutron/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

// StakeTokens funds an account, stakes tokens and returns a stakeID.
func (s *QueryTestSuite) SetupStake(
	addr sdk.AccAddress,
	coins sdk.Coins,
) (stakeID uint64) {
	msgServer := keeper.NewMsgServerImpl(s.App.IncentivesKeeper)
	s.FundAcc(addr, coins)

	msgResponse, err := msgServer.Stake(
		sdk.WrapSDKContext(s.Ctx),
		types.NewMsgSetupStake(addr, coins),
	)
	s.Require().NoError(err)

	return msgResponse.ID
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	pool, _ := s.App.DexKeeper.InitPool(s.Ctx, dextypes.MustNewPairID("tokenA", "tokenB"), 0, 1)
	denom := pool.GetPoolDenom()

	// set up stake with id = 1
	addr := apptesting.SetupAddr(0)
	s.SetupStake(addr, sdk.Coins{sdk.NewCoin(denom, math.NewInt(1000000))})

	s.Commit()
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
