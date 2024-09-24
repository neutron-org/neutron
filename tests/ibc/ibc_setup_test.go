package ibc_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	icstestingutils "github.com/cosmos/interchain-security/v5/testutil/ibc_testing"
	icstestutil "github.com/cosmos/interchain-security/v5/testutil/integration"
	ccvconsumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	ccv "github.com/cosmos/interchain-security/v5/x/ccv/types"
	"github.com/stretchr/testify/suite"

	appparams "github.com/neutron-org/neutron/v4/app/params"
	"github.com/neutron-org/neutron/v4/testutil"
)

var (
	nativeDenom            = appparams.DefaultDenom
	genesisWalletAmount, _ = math.NewIntFromString("10000000000000000000")
)

type IBCTestSuite struct {
	suite.Suite

	coordinator   *ibctesting.Coordinator
	providerChain *ibctesting.TestChain
	providerApp   icstestutil.ProviderApp
	neutronChain  *ibctesting.TestChain // aka chainA
	neutronApp    icstestutil.ConsumerApp
	bundleB       *icstestingutils.ConsumerBundle
	bundleC       *icstestingutils.ConsumerBundle

	neutronCCVPath      *ibctesting.Path
	neutronTransferPath *ibctesting.Path
	neutronChainBPath   *ibctesting.Path
	chainBChainCPath    *ibctesting.Path

	providerAddr           sdk.AccAddress
	neutronAddr            sdk.AccAddress
	providerToNeutronDenom string
}

func TestIBCTestSuite(t *testing.T) {
	suite.Run(t, new(IBCTestSuite))
}

func (s *IBCTestSuite) SetupTest() {
	// we need to redefine this variable to make tests work cause we use untrn as default bond denom in neutron
	sdk.DefaultBondDenom = appparams.DefaultDenom

	// Create coordinator
	s.coordinator = ibctesting.NewCoordinator(s.T(), 0)
	s.providerChain, s.providerApp = icstestingutils.AddProvider[icstestutil.ProviderApp](
		s.T(),
		s.coordinator,
		icstestingutils.ProviderAppIniter,
	)

	// Setup neutron as a consumer chain
	neutronBundle := s.addConsumerChain(testutil.SetupValSetAppIniter, 1)
	s.neutronChain = neutronBundle.Chain
	s.neutronApp = neutronBundle.App
	s.neutronCCVPath = neutronBundle.Path
	s.neutronTransferPath = neutronBundle.TransferPath

	// Setup consumer chainB
	// NOTE: using neutron Setup for chain B, otherwise the consumer chain doesn't have the packetForwarding middleware
	s.bundleB = s.addConsumerChain(testutil.SetupValSetAppIniter, 2)
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
	appIniter icstestingutils.ValSetAppIniter,
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
	consumerGenesisState, found := providerKeeper.GetConsumerGenesis(
		s.providerCtx(),
		bundle.Chain.ChainID,
	)
	s.Require().True(found, "consumer genesis not found")

	genesisState := ccvconsumertypes.GenesisState{
		Params:   consumerGenesisState.Params,
		Provider: consumerGenesisState.Provider,
		NewChain: consumerGenesisState.NewChain,
	}
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

func (s *IBCTestSuite) setupCCVChannel(bundle *icstestingutils.ConsumerBundle) *ibctesting.Path {
	ccvPath := ibctesting.NewPath(bundle.Chain, s.providerChain)

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
) *ibctesting.Path {
	path := ibctesting.NewPath(bundleA.Chain, bundleB.Chain)

	// Set the correct client unbonding period
	clientConfig := ibctesting.NewTendermintConfig()
	clientConfig.UnbondingPeriod = ccv.DefaultConsumerUnbondingPeriod
	clientConfig.TrustingPeriod = clientConfig.UnbondingPeriod * 2 / 3
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
	ccvPath *ibctesting.Path,
	appA icstestutil.ConsumerApp,
	chainA, chainB *ibctesting.TestChain,
) *ibctesting.Path {
	// transfer path will use the same connection as ibc path
	transferPath := ibctesting.NewPath(chainA, chainB)
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
	path *ibctesting.Path,
	sourceEndpoint *ibctesting.Endpoint,
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
	chain *ibctesting.TestChain,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	ctx := chain.GetContext()
	return bk.GetBalance(ctx, addr, denom)
}

func (s *IBCTestSuite) assertBalance(
	bk icstestutil.TestBankKeeper,
	chain *ibctesting.TestChain,
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

//nolint:unused
func (s *IBCTestSuite) assertChainBBalance(addr sdk.AccAddress, denom string, expectedAmt math.Int) {
	s.assertBalance(s.bundleB.App.GetTestBankKeeper(), s.bundleB.Chain, addr, denom, expectedAmt)
}

//nolint:unused
func (s *IBCTestSuite) assertChainCBalance(addr sdk.AccAddress, denom string, expectedAmt math.Int) {
	s.assertBalance(s.bundleC.App.GetTestBankKeeper(), s.bundleC.Chain, addr, denom, expectedAmt)
}
