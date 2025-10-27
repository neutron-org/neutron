package v800_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	v800 "github.com/neutron-org/neutron/v8/app/upgrades/v8.0.0"
	"github.com/neutron-org/neutron/v8/testutil"
	tokenfactorytypes "github.com/neutron-org/neutron/v8/x/tokenfactory/types"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
}

func (suite *UpgradeTestSuite) TestUpgradeDenomsMetadata() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext()

	upgrade := upgradetypes.Plan{
		Name:   v800.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	creator1 := "neutron1lqyjsl7ayetq7z82h0wpk8em057hzpwvrd95tq"
	creator2 := "neutron1mfzr7567u2mp4dnlzx366w35qxwavr64nf0slc"

	denom1, _ := tokenfactorytypes.GetTokenDenom(creator1, "evmos")
	denom2, _ := tokenfactorytypes.GetTokenDenom(creator2, "lamba")

	for _, data := range []struct {
		creator string
		denom   string
	}{
		{
			creator: creator1,
			denom:   denom1,
		},
		{
			creator: creator2,
			denom:   denom2,
		},
	} {
		app.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{
				Denom:    data.denom,
				Exponent: 0,
			}},
			Base: data.denom,
		})

		// add denom from creator
		store := app.TokenFactoryKeeper.GetCreatorPrefixStore(ctx, data.creator)
		store.Set([]byte(data.denom), []byte(data.denom))
	}

	// Apply upgrade
	suite.NoError(app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	// Check fields are correct
	metadata1, _ := app.BankKeeper.GetDenomMetaData(ctx, denom1)
	suite.Require().EqualValues(denom1, metadata1.Name)
	suite.Require().EqualValues(denom1, metadata1.Symbol)
	suite.Require().EqualValues(denom1, metadata1.Display)

	metadata2, _ := app.BankKeeper.GetDenomMetaData(ctx, denom2)
	suite.Require().EqualValues(denom2, metadata2.Name)
	suite.Require().EqualValues(denom2, metadata2.Symbol)
	suite.Require().EqualValues(denom2, metadata2.Display)
}
