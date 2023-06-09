package v102_pion_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/interchain-security/x/ccv/consumer/types"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
	"github.com/stretchr/testify/suite"

	v102 "github.com/neutron-org/neutron/app/upgrades/v1.0.2-pion-1-upgrade"

	"github.com/neutron-org/neutron/testutil"
)

func SetOldCCValidator(ctx sdk.Context, storeKey store.Key, key, bz []byte) {
	store := ctx.KVStore(storeKey)

	store.Set(append([]byte{v102.OldCrossChainValidatorBytePrefix}, key...), bz)
}

func SetOldPendingPackets(ctx sdk.Context, storeKey store.Key, bz []byte) {
	store := ctx.KVStore(storeKey)

	store.Set([]byte{types.PendingDataPacketsByteKey}, bz)
}

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestGlobalFeesUpgrade() {
	var (
		app = suite.GetNeutronZoneApp(suite.ChainA)
		ctx = suite.ChainA.GetContext()
	)

	dataPackets := []ccv.ConsumerPacketData{
		{
			Type: ccv.VscMaturedPacket,
			Data: &ccv.ConsumerPacketData_VscMaturedPacketData{
				VscMaturedPacketData: ccv.NewVSCMaturedPacketData(1),
			},
		},
		{
			Type: ccv.VscMaturedPacket,
			Data: &ccv.ConsumerPacketData_VscMaturedPacketData{
				VscMaturedPacketData: ccv.NewVSCMaturedPacketData(2),
			},
		},
	}
	list := ccv.ConsumerPacketDataList{List: dataPackets}
	bz, err := list.Marshal()
	suite.Require().NoError(err)

	SetOldPendingPackets(ctx, app.GetKey("ccvconsumer"), bz)

	validators := []types.CrossChainValidator{{
		Address: []byte("validator1"),
		Power:   0,
		Pubkey:  nil,
	}, {
		Address: []byte("validator2"),
		Power:   0,
		Pubkey:  nil,
	}}
	for _, validator := range validators {
		bz, err := validator.Marshal()
		suite.Require().NoError(err)

		SetOldCCValidator(ctx, app.GetKey("ccvconsumer"), validator.Address, bz)
	}

	upgrade := upgradetypes.Plan{
		Name:   v102.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	migratedValidators := app.ConsumerKeeper.GetAllCCValidator(ctx)
	suite.Require().Equal(validators, migratedValidators)

	migratedPendingPackets := app.ConsumerKeeper.GetPendingPackets(ctx)
	suite.Require().Equal(list, migratedPendingPackets)
}
