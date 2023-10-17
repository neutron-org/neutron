package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/testutil/apptesting"
	"github.com/neutron-org/neutron/testutil/dex/nullify"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
	"github.com/stretchr/testify/require"
)

// TestIncentivesExportGenesis tests export genesis command for the incentives module.
func (suite *IncentivesTestSuite) TestGenesis() {
	validAddr, _ := apptesting.GenerateTestAddrs()
	genesisState := types.GenesisState{
		Params:      types.DefaultParams(),
		LastStakeId: 10,
		Stakes: []*types.Stake{
			{
				ID:        0,
				Owner:     validAddr,
				StartTime: time.Time{},
				Coins:     sdk.NewCoins(sdk.NewInt64Coin(suite.LPDenom0, 10)),
			},
		},
		LastGaugeId: 10,
		Gauges: []*types.Gauge{
			{
				IsPerpetual: false,
				Coins:       sdk.Coins{sdk.NewInt64Coin("reward", 3000)},
				DistributeTo: types.QueryCondition{
					StartTick: -10,
					EndTick:   10,
					PairID: &dextypes.PairID{
						Token0: "TokenA",
						Token1: "TokenB",
					},
				},
				NumEpochsPaidOver: 1,
				PricingTick:       0,
			},
		},
	}
	suite.App.IncentivesKeeper.InitGenesis(suite.Ctx, genesisState)
	got := suite.App.IncentivesKeeper.ExportGenesis(suite.Ctx)
	require.NotNil(suite.T(), got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(suite.T(), genesisState.Gauges, got.Gauges)
	require.ElementsMatch(suite.T(), genesisState.Stakes, got.Stakes)
	require.Equal(suite.T(), genesisState.LastStakeId, got.LastStakeId)
	require.Equal(suite.T(), genesisState.LastGaugeId, got.LastGaugeId)
}
