package ibc_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	icsibctesting "github.com/cosmos/interchain-security/v3/legacy_ibc_testing/testing"

	icstestingutils "github.com/cosmos/interchain-security/v3/testutil/ibc_testing"
	icstestutil "github.com/cosmos/interchain-security/v3/testutil/integration"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	ccv "github.com/cosmos/interchain-security/v3/x/ccv/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/testutil"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
)

var (
	nativeDenom            = sdk.DefaultBondDenom
	ibcTransferAmount      = math.NewInt(100_000)
	genesisWalletAmount, _ = math.NewIntFromString("10000000000000000000")
)

type IBCTestSuite struct {
	suite.Suite

	coordinator   *icsibctesting.Coordinator
	providerChain *icsibctesting.TestChain
	providerApp   icstestutil.ProviderApp
	neutronChain  *icsibctesting.TestChain // aka chainA
	neutronApp    icstestutil.ConsumerApp
	bundleB       *icstestingutils.ConsumerBundle
	bundleC       *icstestingutils.ConsumerBundle

	neutronCCVPath      *icsibctesting.Path
	neutronTransferPath *icsibctesting.Path
	neutronChainBPath   *icsibctesting.Path
	chainBChainCPath    *icsibctesting.Path

	providerAddr           sdk.AccAddress
	neutronAddr            sdk.AccAddress
	providerToNeutronDenom string
}

func TestIBCTestSuite(t *testing.T) {
	suite.Run(t, new(IBCTestSuite))
}

func (s *IBCTestSuite) SetupTest() {
	// Create coordinator
	s.coordinator = icsibctesting.NewCoordinator(s.T(), 0)
	s.providerChain, s.providerApp = icstestingutils.AddProvider[icstestutil.ProviderApp](
		s.T(),
		s.coordinator,
		icstestingutils.ProviderAppIniter,
	)

	// Setup neutron as a consumer chain
	neutronBundle := s.addConsumerChain(testutil.SetupTestingApp("neutron-1"), 1)
	s.neutronChain = neutronBundle.Chain
	s.neutronApp = neutronBundle.App
	s.neutronCCVPath = neutronBundle.Path
	s.neutronTransferPath = neutronBundle.TransferPath

	// Setup consumer chainB
	// NOTE: using neutron Setup for chain B, otherwise the consumer chain doesn't have the packetForwarding middleware
	s.bundleB = s.addConsumerChain(testutil.SetupTestingApp("chainB-1"), 2)
	// Setup consumer chainC
	s.bundleC = s.addConsumerChain(icstestingutils.ConsumerAppIniter, 3)

	// setup transfer channel between neutron and consumerChainB
	s.neutronChainBPath = s.setupConsumerToConsumerTransferChannel(neutronBundle, s.bundleB)

	// setup transfer channel between consumerChainB and consumerChainC
	s.chainBChainCPath = s.setupConsumerToConsumerTransferChannel(s.bundleB, s.bundleC)

	// Store ibc transfer denom for providerChain=>neutron for test convenience
	fullTransferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronTransferPath.EndpointB.ChannelID,
		nativeDenom,
	)
	transferDenom := transfertypes.ParseDenomTrace(fullTransferDenomPath).IBCDenom()
	s.providerToNeutronDenom = transferDenom

	// Store default addresses from neutron and provider chain for test convenience
	s.providerAddr = s.providerChain.SenderAccount.GetAddress()
	s.neutronAddr = s.neutronChain.SenderAccount.GetAddress()

	// ensure genesis balances are as expected
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, genesisWalletAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount)
}

func (s *IBCTestSuite) addConsumerChain(
	appIniter icsibctesting.AppIniter,
	chainIdx int,
) *icstestingutils.ConsumerBundle {
	bundle := icstestingutils.AddConsumer[icstestutil.ProviderApp, icstestutil.ConsumerApp](
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
	consumerKeeper.SetProviderChannel(bundle.GetCtx(), bundle.Path.EndpointA.ChannelID)

	return bundle
}

func (s *IBCTestSuite) setupCCVChannel(bundle *icstestingutils.ConsumerBundle) *icsibctesting.Path {
	ccvPath := icsibctesting.NewPath(bundle.Chain, s.providerChain)

	providerKeeper := s.providerApp.GetProviderKeeper()
	neutronKeeper := bundle.GetKeeper()

	providerEndpointClientID, found := providerKeeper.GetConsumerClientId(
		s.providerCtx(),
		bundle.Chain.ChainID,
	)
	s.Require().True(found, "provider endpoint clientID not found")
	ccvPath.EndpointB.ClientID = providerEndpointClientID

	consumerEndpointClientID, found := neutronKeeper.GetProviderClientID(bundle.GetCtx())
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
	appA icstestutil.ConsumerApp,
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

func (s *IBCTestSuite) providerCtx() sdk.Context {
	return s.providerChain.GetContext()
}

// Helper Methods /////////////////////////////////////////////////////////////

func (s *IBCTestSuite) IBCTransfer(
	path *icsibctesting.Path,
	sourceEndpoint *icsibctesting.Endpoint,
	fromAddr sdk.AccAddress,
	toAddr sdk.AccAddress,
	transferDenom string,
	transferAmount math.Int,
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

	// Relay transfer msg to Neutron chain
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	//nolint:errcheck // this will return an error for multi-hop routes; that's expected
	path.RelayPacket(packet)
}

func (s *IBCTestSuite) IBCTransferProviderToNeutron(
	providerAddr sdk.AccAddress,
	neutronAddr sdk.AccAddress,
	transferDenom string,
	transferAmount math.Int,
	memo string,
) {
	s.IBCTransfer(
		s.neutronTransferPath,
		s.neutronTransferPath.EndpointB,
		providerAddr,
		neutronAddr,
		transferDenom,
		transferAmount,
		memo,
	)
}

func (s *IBCTestSuite) getBalance(
	bk icstestutil.TestBankKeeper,
	chain *icsibctesting.TestChain,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	ctx := chain.GetContext()
	return bk.GetBalance(ctx, addr, denom)
}

func (s *IBCTestSuite) assertBalance(
	bk icstestutil.TestBankKeeper,
	chain *icsibctesting.TestChain,
	addr sdk.AccAddress,
	denom string,
	expectedAmt math.Int,
) {
	actualAmt := s.getBalance(bk, chain, addr, denom).Amount
	s.Assert().
		Equal(expectedAmt, actualAmt, "Expected amount of %s: %s; Got: %s", denom, expectedAmt, actualAmt)
}

func (s *IBCTestSuite) assertNeutronBalance(
	addr sdk.AccAddress,
	denom string,
	expectedAmt math.Int,
) {
	s.assertBalance(s.neutronApp.GetTestBankKeeper(), s.neutronChain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) assertProviderBalance(
	addr sdk.AccAddress,
	denom string,
	expectedAmt math.Int,
) {
	s.assertBalance(s.providerApp.GetTestBankKeeper(), s.providerChain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) assertChainBBalance(addr sdk.AccAddress, denom string, expectedAmt math.Int) {
	s.assertBalance(s.bundleB.App.GetTestBankKeeper(), s.bundleB.Chain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) assertChainCBalance(addr sdk.AccAddress, denom string, expectedAmt math.Int) {
	s.assertBalance(s.bundleC.App.GetTestBankKeeper(), s.bundleC.Chain, addr, denom, expectedAmt)
}

func (s *IBCTestSuite) ReceiverOverrideAddr(channel, sender string) sdk.AccAddress {
	addr, err := packetforward.GetReceiver(channel, sender)
	if err != nil {
		panic("Cannot calc receiver override: " + err.Error())
	}
	return sdk.MustAccAddressFromBech32(addr)
}

//nolint:unparam // keep this flexible even if we aren't currently using all the params
func (s *IBCTestSuite) neutronDeposit(
	token0 string,
	token1 string,
	depositAmount0 math.Int,
	depositAmount1 math.Int,
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
		[]math.Int{depositAmount0},
		[]math.Int{depositAmount1},
		[]int64{tickIndex},
		[]uint64{fee},
		[]*dextypes.DepositOptions{{DisableAutoswap: false}},
	)

	// execute deposit msg
	_, err := s.neutronChain.SendMsgs(msgDeposit)
	s.Assert().NoError(err, "Deposit Failed")
}

func (s *IBCTestSuite) RelayAllPacketsAToB(path *icsibctesting.Path) error {
	sentPackets := path.EndpointA.Chain.SentPackets
	chainB := path.EndpointB.Chain
	if len(sentPackets) == 0 {
		return fmt.Errorf("No packets to send")
	}

	for _, packet := range sentPackets {
		// Skip if packet has already been sent
		ack, _ := chainB.App.GetIBCKeeper().ChannelKeeper.
			GetPacketAcknowledgement(chainB.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		if ack != nil {
			continue
		}
		err := path.RelayPacket(packet)
		if err != nil {
			return err
		}
	}
	return nil
}
