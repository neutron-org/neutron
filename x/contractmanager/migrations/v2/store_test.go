package v2_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/testutil"
	v2 "github.com/neutron-org/neutron/x/contractmanager/migrations/v2"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	typesv1 "github.com/neutron-org/neutron/x/contractmanager/types/v1"
	"github.com/stretchr/testify/suite"
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
	suite.NoError(v2.MigrateStore(ctx, storeKey, cdc))

	// Check elements migrated properly
	suite.Require().ElementsMatch(app.ContractManagerKeeper.GetAllFailures(ctx), []types.Failure{
		{
			Address: addressOne,
			Id:      0,
			AckType: types.Ack,
			Packet:  nil,
			Ack:     nil,
		},
		{
			Address: addressOne,
			Id:      1,
			AckType: types.Ack,
			Packet:  nil,
			Ack:     nil,
		},
		{
			Address: addressTwo,
			Id:      0,
			AckType: types.Ack,
			Packet:  nil,
			Ack:     nil,
		},
		{
			Address: addressTwo,
			Id:      1,
			AckType: types.Ack,
			Packet:  nil,
			Ack:     nil,
		},
	})

	// Check getting element works
	failure, err := app.ContractManagerKeeper.GetFailure(ctx, sdk.MustAccAddressFromBech32(addressTwo), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(failure, &types.Failure{
		Address: addressTwo,
		Id:      1,
		AckType: types.Ack,
		Packet:  nil,
		Ack:     nil,
	})

	// Non-existent returns error
	_, err = app.ContractManagerKeeper.GetFailure(ctx, sdk.MustAccAddressFromBech32(addressTwo), 2)
	suite.Require().Error(err)

	// Check next id key is correct
	oneKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressOne)
	suite.Require().Equal(oneKey, uint64(2))
	twoKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressTwo)
	suite.Require().Equal(twoKey, uint64(2))
}
