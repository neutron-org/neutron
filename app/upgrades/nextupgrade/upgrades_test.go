package nextupgrade_test

import (
	"testing"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/neutron-org/neutron/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/testutil"
	"github.com/stretchr/testify/suite"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestGlobalFeesUpgrade() {
	var (
		app               = suite.GetNeutronZoneApp(suite.ChainA)
		globalFeeSubspace = app.GetSubspace(globalfee.ModuleName)
		ctx               = suite.ChainA.GetContext()
	)

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage))

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes))
	suite.Require().True(globalFeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage))

	var globalMinGasPrices sdk.DecCoins
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	requiredGlobalFees := sdk.DecCoins{
		sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.MustNewDecFromStr("0.026")),
		sdk.NewDecCoinFromDec("ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349", sdk.MustNewDecFromStr("0.25")),
		sdk.NewDecCoinFromDec("untrn", sdk.MustNewDecFromStr("0.9")),
	}
	suite.Require().Equal(requiredGlobalFees, globalMinGasPrices)

	var actualBypassFeeMessages []string
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes, &actualBypassFeeMessages)
	requiredBypassMinFeeMsgTypes := []string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
	}
	suite.Require().Equal(requiredBypassMinFeeMsgTypes, actualBypassFeeMessages)

	var actualTotalBypassMinFeeMsgGasUsage uint64
	globalFeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &actualTotalBypassMinFeeMsgGasUsage)
	requiredTotalBypassMinFeeMsgGasUsage := uint64(1_000_000)
	suite.Require().Equal(requiredTotalBypassMinFeeMsgGasUsage, actualTotalBypassMinFeeMsgGasUsage)
}

func (suite *UpgradeTestSuite) TestFailuresUpgrade() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(contractmanagertypes.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	addressOne := testutil.TestOwnerAddress
	addressTwo := "neutron1fxudpred77a0grgh69u0j7y84yks5ev4n5050z45kecz792jnd6scqu98z"

	store := ctx.KVStore(storeKey)
	var i uint64
	for i = 0; i < 4; i++ {
		var addr string
		if i < 2 {
			addr = addressOne
		} else {
			addr = addressTwo
		}
		failure := contractmanagertypes.OldFailure{
			ChannelId: "channel-0",
			Address:   addr,
			Id:        i % 2,
			AckType:   contractmanagertypes.Ack,
		}
		bz := cdc.MustMarshal(&failure)
		store.Set(contractmanagertypes.GetFailureKey(failure.Address, failure.Id), bz)
	}

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}
	app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade)

	// check elements migrated properly
	suite.Require().ElementsMatch(app.ContractManagerKeeper.GetAllFailures(ctx), []contractmanagertypes.Failure{
		{
			Address: addressOne,
			Id:      0,
			AckType: contractmanagertypes.Ack,
			Packet:  nil,
			Ack:     nil,
		},
		{
			Address: addressOne,
			Id:      1,
			AckType: contractmanagertypes.Ack,
			Packet:  nil,
			Ack:     nil,
		},
		{
			Address: addressTwo,
			Id:      0,
			AckType: contractmanagertypes.Ack,
			Packet:  nil,
			Ack:     nil,
		},
		{
			Address: addressTwo,
			Id:      1,
			AckType: contractmanagertypes.Ack,
			Packet:  nil,
			Ack:     nil,
		},
	})

	// check getting element works
	failure, err := app.ContractManagerKeeper.GetFailure(ctx, sdk.MustAccAddressFromBech32(addressTwo), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(failure, &contractmanagertypes.Failure{
		Address: addressTwo,
		Id:      1,
		AckType: contractmanagertypes.Ack,
		Packet:  nil,
		Ack:     nil,
	})

	// non-existent returns error
	_, err = app.ContractManagerKeeper.GetFailure(ctx, sdk.MustAccAddressFromBech32(addressTwo), 2)
	suite.Require().Error(err)

	// check id's is in order
	oneKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressOne)
	suite.Require().Equal(oneKey, uint64(2))
	twoKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressTwo)
	suite.Require().Equal(twoKey, uint64(2))
}
