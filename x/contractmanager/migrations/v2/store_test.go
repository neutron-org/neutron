package v2_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/testutil"
	v2 "github.com/neutron-org/neutron/v6/x/contractmanager/migrations/v2"
	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
	typesv1 "github.com/neutron-org/neutron/v6/x/contractmanager/types/v1"
)

type V2ContractManagerMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V2ContractManagerMigrationTestSuite))
}

func (suite *V2ContractManagerMigrationTestSuite) TestFailuresUpgrade() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	addressOne := testutil.TestOwnerAddress
	addressTwo := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"

	// Write old state
	store := ctx.KVStore(storeKey)
	var i uint64
	for i = 0; i < 4; i++ {
		var addr string
		if i < 2 {
			addr = addressOne
		} else {
			addr = addressTwo
		}
		failure := typesv1.Failure{
			ChannelId: "channel-0",
			Address:   addr,
			Id:        i % 2,
			AckType:   types.Ack,
		}
		bz := cdc.MustMarshal(&failure)
		store.Set(types.GetFailureKey(failure.Address, failure.Id), bz)
	}

	// Run migration
	suite.NoError(v2.MigrateStore(ctx, storeKey))

	// Check elements should be empty
	expected := app.ContractManagerKeeper.GetAllFailures(ctx)
	suite.Require().ElementsMatch(expected, []types.Failure{})

	// Non-existent returns error
	_, err := app.ContractManagerKeeper.GetFailure(ctx, sdk.MustAccAddressFromBech32(addressTwo), 0)
	suite.Require().Error(err)

	// Check next id key is reset
	oneKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressOne)
	suite.Require().Equal(oneKey, uint64(0))
	twoKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressTwo)
	suite.Require().Equal(twoKey, uint64(0))
}
