package v3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v7/testutil"
	v3 "github.com/neutron-org/neutron/v7/x/tokenfactory/migrations/v3"
)

type V3TokenfactoryMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V3TokenfactoryMigrationTestSuite))
}

func (suite *V3TokenfactoryMigrationTestSuite) TestParamsUpgrade() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext()
	)

	// Write denoms metadata
	denom1 := "factory/neutron1lqyjsl7ayetq7z82h0wpk8em057hzpwvrd95tq/evmos"
	denom2 := "factory/neutron1mfzr7567u2mp4dnlzx366w35qxwavr64nf0slc/lalala"

	for _, denom := range []string{denom1, denom2} {
		app.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{
				Denom:    denom,
				Exponent: 0,
			}},
			Base:    denom,
			Name:    denom,
			Symbol:  denom,
			Display: denom,
		})
	}

	// Run migration
	suite.NoError(v3.MigrateStore(ctx, app.TokenFactoryKeeper, app.BankKeeper))

	// Check fields are correct
	metadata1, _ := app.BankKeeper.GetDenomMetaData(ctx, denom1)
	metadata2, _ := app.BankKeeper.GetDenomMetaData(ctx, denom2)

	suite.Require().EqualValues(denom1, metadata1.Name)
	suite.Require().EqualValues(denom1, metadata1.Symbol)
	suite.Require().EqualValues(denom1, metadata1.Display)
	suite.Require().EqualValues(denom2, metadata2.Name)
	suite.Require().EqualValues(denom2, metadata2.Symbol)
	suite.Require().EqualValues(denom2, metadata2.Display)
}
