package ibc_test

import (
	"encoding/json"
	"fmt"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	icsibctesting "github.com/cosmos/interchain-security/v3/legacy_ibc_testing/testing"

	"github.com/cometbft/cometbft/libs/log"
	icstestingutils "github.com/cosmos/interchain-security/v3/testutil/ibc_testing"
	testutil "github.com/cosmos/interchain-security/v3/testutil/integration"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	ccv "github.com/cosmos/interchain-security/v3/x/ccv/types"
	app "github.com/neutron-org/neutron/app"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/suite"
)

var (
	nativeDenom            = sdk.DefaultBondDenom
	ibcTransferAmount      = sdk.NewInt(100_000)
	genesisWalletAmount, _ = sdk.NewIntFromString("10000000000000000000")
)

type IBCTestSuite struct {
	suite.Suite

	coordinator   *icsibctesting.Coordinator
	providerChain *icsibctesting.TestChain
	providerApp   testutil.ProviderApp
	dualityChain  *icsibctesting.TestChain // aka chainA
	dualityApp    testutil.ConsumerApp
	bundleB       *icstestingutils.ConsumerBundle
	bundleC       *icstestingutils.ConsumerBundle

	dualityCCVPath      *icsibctesting.Path
	dualityTransferPath *icsibctesting.Path
	dualityChainBPath   *icsibctesting.Path
	chainBChainCPath    *icsibctesting.Path

	providerAddr           sdk.AccAddress
	dualityAddr            sdk.AccAddress
	providerToDualityDenom string
}

func TestIBCTestSuite(t *testing.T) {
	suite.Run(t, new(IBCTestSuite))
}

func (s *IBCTestSuite) SetupTest() {
	// Create coordinator
	s.coordinator = icsibctesting.NewCoordinator(s.T(), 0)
	s.providerChain, s.providerApp = icstestingutils.AddProvider[testutil.ProviderApp](
		s.T(),
		s.coordinator,
		icstestingutils.ProviderAppIniter,
	)

	// Setup duality as a consumer chain
	dualityBundle := s.addConsumerChain(dualityAppIniter, 1)
	s.dualityChain = dualityBundle.Chain
	s.dualityApp = dualityBundle.App
	s.dualityCCVPath = dualityBundle.Path
	s.dualityTransferPath = dualityBundle.TransferPath

	// Setup consumer chainB
	// NOTE: using dualityAppIniter otherwise the consumer chain doesn't have the packetForwarding middleware
	s.bundleB = s.addConsumerChain(dualityAppIniter, 2)
	// Setup consumer chainC
	s.bundleC = s.addConsumerChain(icstestingutils.ConsumerAppIniter, 3)

	// setup transfer channel between duality and consumerChainB
	s.dualityChainBPath = s.setupConsumerToConsumerTransferChannel(dualityBundle, s.bundleB)

	// setup transfer channel between consumerChainB and consumerChainC
	s.chainBChainCPath = s.setupConsumerToConsumerTransferChannel(s.bundleB, s.bundleC)

	// Store ibc transfer denom for providerChain=>duality for test convenience
	fullTransferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.dualityTransferPath.EndpointB.ChannelID,
		nativeDenom,
	)
	transferDenom := transfertypes.ParseDenomTrace(fullTransferDenomPath).IBCDenom()
	s.providerToDualityDenom = transferDenom

	// Store default addresses from duality and provider chain for test convenience
	s.providerAddr = s.providerChain.SenderAccount.GetAddress()
	s.dualityAddr = s.dualityChain.SenderAccount.GetAddress()

	// ensure genesis balances are as expected
	s.assertDualityBalance(s.dualityAddr, nativeDenom, genesisWalletAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount)
}

func (s *IBCTestSuite) addConsumerChain(
	appIniter icsibctesting.AppIniter,
	chainIdx int,
) *icstestingutils.ConsumerBundle {
	bundle := icstestingutils.AddConsumer[testutil.ProviderApp, testutil.ConsumerApp](
		s.coordinator,
		&s.Suite,
		chainIdx,
		appIniter,
	)
	providerKeeper := s.providerApp.GetProviderKeeper()
	consumerKeeper := bundle.GetKeeper()
	genesisState, found := providerKeeper.GetConsumerGenesis(
		s.providerCtx(),
		bundle.Chain.ChainID,
	)
	s.Require().True(found, "consumer genesis not found")
	consumerKeeper.InitGenesis(bundle.GetCtx(), &genesisState)
	bundle.Path = s.setupCCVChannel(bundle)
	bundle.TransferPath = s.setupTransferChannel(
		bundle.Path,
		bundle.App,
		bundle.Chain,
		s.providerChain,
	)

	return bundle
}

func (s *IBCTestSuite) setupCCVChannel(bundle *icstestingutils.ConsumerBundle) *icsibctesting.Path {
	ccvPath := icsibctesting.NewPath(bundle.Chain, s.providerChain)

	providerKeeper := s.providerApp.GetProviderKeeper()
	dualityKeeper := bundle.GetKeeper()

	providerEndpointClientID, found := providerKeeper.GetConsumerClientId(
		s.providerCtx(),
		bundle.Chain.ChainID,
	)
	s.Require().True(found, "provider endpoint clientID not found")
	ccvPath.EndpointB.ClientID = providerEndpointClientID

	consumerEndpointClientID, found := dualityKeeper.GetProviderClientID(bundle.GetCtx())
	s.Require().True(found, "consumer endpoint clientID not found")
	ccvPath.EndpointA.ClientID = consumerEndpointClientID

	ccvPath.EndpointA.ChannelConfig.PortID = ccv.ConsumerPortID
	ccvPath.EndpointB.ChannelConfig.PortID = ccv.ProviderPortID
	ccvPath.EndpointA.ChannelConfig.Version = ccv.Version
	ccvPath.EndpointB.ChannelConfig.Version = ccv.Version
	ccvPath.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	ccvPath.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	// Create ccv connection
	s.coordinator.CreateConnections(ccvPath)

	// create ccv channel
	s.coordinator.CreateChannels(ccvPath)

	return ccvPath
}

func (s *IBCTestSuite) setupConsumerToConsumerTransferChannel(
	bundleA, bundleB *icstestingutils.ConsumerBundle,
) *icsibctesting.Path {
	path := icsibctesting.NewPath(bundleA.Chain, bundleB.Chain)

	// Set the correct client unbonding period
	clientConfig := icsibctesting.NewTendermintConfig()
	clientConfig.UnbondingPeriod = ccvconsumertypes.DefaultConsumerUnbondingPeriod
	path.EndpointA.ClientConfig = clientConfig
	path.EndpointB.ClientConfig = clientConfig

	path.EndpointA.ChannelConfig.PortID = "transfer"
	path.EndpointB.ChannelConfig.PortID = "transfer"
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	s.coordinator.Setup(path)

	return path
}

func (s *IBCTestSuite) setupTransferChannel(
	ccvPath *icsibctesting.Path,
	appA testutil.ConsumerApp,
	chainA, chainB *icsibctesting.TestChain,
) *icsibctesting.Path {
	// transfer path will use the same connection as ibc path
	transferPath := icsibctesting.NewPath(chainA, chainB)
	transferPath.EndpointA.ChannelConfig.PortID = transfertypes.PortID
	transferPath.EndpointB.ChannelConfig.PortID = transfertypes.PortID
	transferPath.EndpointA.ChannelConfig.Version = transfertypes.Version
	transferPath.EndpointB.ChannelConfig.Version = transfertypes.Version
	transferPath.EndpointA.ClientID = ccvPath.EndpointA.ClientID
	transferPath.EndpointA.ConnectionID = ccvPath.EndpointA.ConnectionID
	transferPath.EndpointB.ClientID = ccvPath.EndpointB.ClientID
	transferPath.EndpointB.ConnectionID = ccvPath.EndpointB.ConnectionID

	// IBC channel handshake will automatically initiate transfer channel handshake on ACK
	// so transfer channel will be on stage INIT when CompleteSetupIBCChannel returns
	destinationChannelID := appA.GetConsumerKeeper().
		GetDistributionTransmissionChannel(chainA.GetContext())
	transferPath.EndpointA.ChannelID = destinationChannelID

	// Complete TRY, ACK, CONFIRM for transfer path
	err := transferPath.EndpointB.ChanOpenTry()
	s.Require().NoError(err)

	err = transferPath.EndpointA.ChanOpenAck()
	s.Require().NoError(err)

	err = transferPath.EndpointB.ChanOpenConfirm()
	s.Require().NoError(err)

	// ensure counterparty is up to date
	err = transferPath.EndpointA.UpdateClient()
	s.Require().NoError(err)

	return transferPath
}

func dualityAppIniter() (icsibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := app.MakeEncodingConfig()
	db := dbm.NewMemDB()
	testApp := app.New(
		log.NewNopLogger(),
		"neutron-1",
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		sims.EmptyAppOptions{},
		nil,
	)

	return testApp, app.NewDefaultGenesisState(testApp.AppCodec())
}

func (s *IBCTestSuite) providerCtx() sdk.Context {
	return s.providerChain.GetContext()
}

func (s *IBCTestSuite) dualityCtx() sdk.Context {
	return s.dualityChain.GetContext()
}

// Helper Methods /////////////////////////////////////////////////////////////

func (s *IBCTestSuite) IBCTransfer(
	path *icsibctesting.Path,
	sourceEndpoint *icsibctesting.Endpoint,
	fromAddr sdk.AccAddress,
	toAddr sdk.AccAddress,
	transferDenom string,
	transferAmount sdk.Int,
	memo string,
) {
	timeoutHeight := clienttypes.NewHeight(1, 110)

	// Create Transfer Msg
	transferMsg := transfertypes.NewMsgTransfer(sourceEndpoint.ChannelConfig.PortID,
		sourceEndpoint.ChannelID,
		sdk.NewCoin(transferDenom, transferAmount),
		fromAddr.String(),
		toAddr.String(),
		timeoutHeight,
		0,
		memo,
	)

	// Send message from provider chain
	res, err := sourceEndpoint.Chain.SendMsgs(transferMsg)
	s.Assert().NoError(err)

	// Relay transfer msg to Duality chain
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	path.RelayPacket(packet)
	s.Assert().NoError(err)
}

func (s *IBCTestSuite) IBCTransferProviderToDuality(
	providerAddr sdk.AccAddress,
	dualityAddr sdk.AccAddress,
	transferDenom string,
	transferAmount sdk.Int,
	memo string,
) {
	s.IBCTransfer(
		s.dualityTransferPath,
		s.dualityTransferPath.EndpointB,
		providerAddr,
		dualityAddr,
		transferDenom,
		transferAmount,
		memo,
	)
}

func (s *IBCTestSuite) getBalance(
	bk testutil.TestBankKeeper,
	chain *icsibctesting.TestChain,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	ctx := chain.GetContext()
	return bk.GetBalance(ctx, addr, denom)
}

func (s *IBCTestSuite) assertBalance(
	bk testutil.TestBankKeeper,
	chain *icsibctesting.TestChain,
	addr sdk.AccAddress,
	denom string,
	expectedAmt sdk.Int,
) {
	actualAmt := s.getBalance(bk, chain, addr, denom).Amount
	s.Assert().
		Equal(expectedAmt, actualAmt, "Expected amount of %s: %s; Got: %s", denom, expectedAmt, actualAmt)
}

func (s *IBCTestSuite) assertDualityBalance(
	addr sdk.AccAddress,
	denom string,
	expectedAmt sdk.Int,
) {
	s.assertBalance(s.dualityApp.GetTestBankKeeper(), s.dualityChain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) assertProviderBalance(
	addr sdk.AccAddress,
	denom string,
	expectedAmt sdk.Int,
) {
	s.assertBalance(s.providerApp.GetTestBankKeeper(), s.providerChain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) assertChainBBalance(addr sdk.AccAddress, denom string, expectedAmt sdk.Int) {
	s.assertBalance(s.bundleB.App.GetTestBankKeeper(), s.bundleB.Chain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) assertChainCBalance(addr sdk.AccAddress, denom string, expectedAmt sdk.Int) {
	s.assertBalance(s.bundleC.App.GetTestBankKeeper(), s.bundleC.Chain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) dualityDeposit(
	token0 string,
	token1 string,
	depositAmount0 sdk.Int,
	depositAmount1 sdk.Int,
	tickIndex int64,
	fee uint64,
	creator sdk.AccAddress,
) {
	// create deposit msg
	msgDeposit := dextypes.NewMsgDeposit(
		creator.String(),
		creator.String(),
		token0,
		token1,
		[]sdk.Int{depositAmount0},
		[]sdk.Int{depositAmount1},
		[]int64{tickIndex},
		[]uint64{fee},
		[]*dextypes.DepositOptions{{false}},
	)

	// execute deposit msg
	_, err := s.dualityChain.SendMsgs(msgDeposit)
	s.Assert().NoError(err, "Deposit Failed")
}

func (s *IBCTestSuite) RelayAllPacketsAToB(path *icsibctesting.Path) error {
	sentPackets := path.EndpointA.Chain.SentPackets
	if len(sentPackets) == 0 {
		return fmt.Errorf("No packets to send")
	}

	for _, packet := range sentPackets {
		err := path.RelayPacket(packet)
		if err != nil {
			return err
		}
	}
	return nil
}
