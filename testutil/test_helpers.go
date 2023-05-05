package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/interchain-security/legacy_ibc_testing/testing"
	icssimapp "github.com/cosmos/interchain-security/testutil/ibc_testing"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"

	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	appProvider "github.com/cosmos/interchain-security/app/provider"
	e2e "github.com/cosmos/interchain-security/testutil/e2e"
	ccvutils "github.com/cosmos/interchain-security/x/ccv/utils"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/neutron-org/neutron/app"
	ictxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"

	consumertypes "github.com/cosmos/interchain-security/x/ccv/consumer/types"
	providertypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
)

var (
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"

	TestInterchainID = "owner_id"

	// provider-consumer connection takes connection-0
	ConnectionOne = "connection-1"

	// TestVersion defines a reusable interchainaccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ConnectionOne,
		HostConnectionId:       ConnectionOne,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

func init() {
	ibctesting.DefaultTestingAppInit = SetupTestingApp
	app.GetDefaultConfig()
}

type IBCConnectionTestSuite struct {
	suite.Suite
	Coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	ChainProvider *ibctesting.TestChain
	ChainA        *ibctesting.TestChain
	ChainB        *ibctesting.TestChain

	ProviderApp e2e.ProviderApp
	ChainAApp   e2e.ConsumerApp
	ChainBApp   e2e.ConsumerApp

	CCVPathA     *ibctesting.Path
	CCVPathB     *ibctesting.Path
	Path         *ibctesting.Path
	TransferPath *ibctesting.Path
}

func GetTestConsumerAdditionProp(chain *ibctesting.TestChain) *providertypes.ConsumerAdditionProposal {
	prop := providertypes.NewConsumerAdditionProposal(
		chain.ChainID,
		"description",
		chain.ChainID,
		chain.LastHeader.GetHeight().(clienttypes.Height),
		[]byte("gen_hash"),
		[]byte("bin_hash"),
		time.Now(),
		consumertypes.DefaultConsumerRedistributeFrac,
		consumertypes.DefaultBlocksPerDistributionTransmission,
		consumertypes.DefaultHistoricalEntries,
		ccv.DefaultCCVTimeoutPeriod,
		consumertypes.DefaultTransferTimeoutPeriod,
		consumertypes.DefaultConsumerUnbondingPeriod,
	).(*providertypes.ConsumerAdditionProposal)

	return prop
}

func (suite *IBCConnectionTestSuite) SetupTest() {
	suite.Coordinator = NewProviderConsumerCoordinator(suite.T())
	suite.ChainProvider = suite.Coordinator.GetChain(ibctesting.GetChainID(1))
	suite.ChainA = suite.Coordinator.GetChain(ibctesting.GetChainID(2))
	suite.ChainB = suite.Coordinator.GetChain(ibctesting.GetChainID(3))
	suite.ProviderApp = suite.ChainProvider.App.(*appProvider.App)
	suite.ChainAApp = suite.ChainA.App.(*app.App)
	suite.ChainBApp = suite.ChainB.App.(*app.App)

	providerKeeper := suite.ProviderApp.GetProviderKeeper()
	consumerKeeperA := suite.ChainAApp.GetConsumerKeeper()
	consumerKeeperB := suite.ChainBApp.GetConsumerKeeper()

	// valsets must match
	providerValUpdates := tmtypes.TM2PB.ValidatorUpdates(suite.ChainProvider.Vals)
	consumerAValUpdates := tmtypes.TM2PB.ValidatorUpdates(suite.ChainA.Vals)
	consumerBValUpdates := tmtypes.TM2PB.ValidatorUpdates(suite.ChainB.Vals)
	suite.Require().True(len(providerValUpdates) == len(consumerAValUpdates), "initial valset not matching")
	suite.Require().True(len(providerValUpdates) == len(consumerBValUpdates), "initial valset not matching")

	for i := 0; i < len(providerValUpdates); i++ {
		addr1, _ := ccvutils.TMCryptoPublicKeyToConsAddr(providerValUpdates[i].PubKey)
		addr2, _ := ccvutils.TMCryptoPublicKeyToConsAddr(consumerAValUpdates[i].PubKey)
		addr3, _ := ccvutils.TMCryptoPublicKeyToConsAddr(consumerBValUpdates[i].PubKey)
		suite.Require().True(bytes.Equal(addr1, addr2), "validator mismatch")
		suite.Require().True(bytes.Equal(addr1, addr3), "validator mismatch")
	}

	// move chains to the next block
	suite.ChainProvider.NextBlock()
	suite.ChainA.NextBlock()
	suite.ChainB.NextBlock()

	// create consumer client on provider chain and set as consumer client for consumer chainID in provider keeper.
	prop1 := GetTestConsumerAdditionProp(suite.ChainA)
	err := providerKeeper.CreateConsumerClient(
		suite.ChainProvider.GetContext(),
		prop1,
	)
	suite.Require().NoError(err)

	prop2 := GetTestConsumerAdditionProp(suite.ChainB)
	err = providerKeeper.CreateConsumerClient(
		suite.ChainProvider.GetContext(),
		prop2,
	)
	suite.Require().NoError(err)

	// move provider to next block to commit the state
	suite.ChainProvider.NextBlock()

	// initialize the consumer chain with the genesis state stored on the provider
	consumerGenesisA, found := providerKeeper.GetConsumerGenesis(
		suite.ChainProvider.GetContext(),
		suite.ChainA.ChainID,
	)
	suite.Require().True(found, "consumer genesis not found")
	consumerKeeperA.InitGenesis(suite.ChainA.GetContext(), &consumerGenesisA)

	// initialize the consumer chain with the genesis state stored on the provider
	consumerGenesisB, found := providerKeeper.GetConsumerGenesis(
		suite.ChainProvider.GetContext(),
		suite.ChainB.ChainID,
	)
	suite.Require().True(found, "consumer genesis not found")
	consumerKeeperB.InitGenesis(suite.ChainB.GetContext(), &consumerGenesisB)

	// create paths for the CCV channel
	suite.CCVPathA = ibctesting.NewPath(suite.ChainA, suite.ChainProvider)
	suite.CCVPathB = ibctesting.NewPath(suite.ChainB, suite.ChainProvider)
	SetupCCVPath(suite.CCVPathA, suite)
	SetupCCVPath(suite.CCVPathB, suite)

	suite.SetupCCVChannels()

	suite.Path = NewICAPath(suite.ChainA, suite.ChainB, suite.ChainProvider)

	suite.Coordinator.SetupConnections(suite.Path)
}

func (suite *IBCConnectionTestSuite) ConfigureTransferChannel() {
	suite.TransferPath = NewTransferPath(suite.ChainA, suite.ChainB, suite.ChainProvider)
	suite.Coordinator.SetupConnections(suite.TransferPath)
	err := SetupTransferPath(suite.TransferPath)
	suite.Require().NoError(err)
}

func (suite *IBCConnectionTestSuite) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.MintCoins(suite.ChainA.GetContext(), tokenfactorytypes.ModuleName, amounts)
	suite.Require().NoError(err)

	err = bankKeeper.SendCoinsFromModuleToAccount(suite.ChainA.GetContext(), tokenfactorytypes.ModuleName, acc, amounts)
	suite.Require().NoError(err)
}

// update CCV path with correct info
func SetupCCVPath(path *ibctesting.Path, suite *IBCConnectionTestSuite) {
	// - set provider endpoint's clientID
	consumerClient, found := suite.ProviderApp.GetProviderKeeper().GetConsumerClientId(
		suite.ChainProvider.GetContext(),
		path.EndpointA.Chain.ChainID,
	)

	suite.Require().True(found, "consumer client not found")
	path.EndpointB.ClientID = consumerClient

	// - set consumer endpoint's clientID
	consumerKeeper := path.EndpointA.Chain.App.(*app.App).GetConsumerKeeper()
	providerClient, found := consumerKeeper.GetProviderClientID(path.EndpointA.Chain.GetContext())
	suite.Require().True(found, "provider client not found")
	path.EndpointA.ClientID = providerClient

	// - client config
	trustingPeriodFraction := suite.ProviderApp.GetProviderKeeper().GetTrustingPeriodFraction(suite.ChainProvider.GetContext())

	providerUnbondingPeriod := suite.ProviderApp.GetStakingKeeper().UnbondingTime(suite.ChainProvider.GetContext())
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = providerUnbondingPeriod
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccv.CalculateTrustPeriod(providerUnbondingPeriod, trustingPeriodFraction)
	consumerUnbondingPeriod := consumerKeeper.GetUnbondingPeriod(path.EndpointA.Chain.GetContext())
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriod
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccv.CalculateTrustPeriod(consumerUnbondingPeriod, trustingPeriodFraction)
	// - channel config
	path.EndpointA.ChannelConfig.PortID = ccv.ConsumerPortID
	path.EndpointB.ChannelConfig.PortID = ccv.ProviderPortID
	path.EndpointA.ChannelConfig.Version = ccv.Version
	path.EndpointB.ChannelConfig.Version = ccv.Version
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
}

func (suite *IBCConnectionTestSuite) SetupCCVChannels() {
	paths := []*ibctesting.Path{suite.CCVPathA, suite.CCVPathB}
	for _, path := range paths {
		suite.Coordinator.CreateConnections(path)

		err := path.EndpointA.ChanOpenInit()
		suite.Require().NoError(err)

		err = path.EndpointB.ChanOpenTry()
		suite.Require().NoError(err)

		err = path.EndpointA.ChanOpenAck()
		suite.Require().NoError(err)

		err = path.EndpointB.ChanOpenConfirm()
		suite.Require().NoError(err)

		err = path.EndpointA.UpdateClient()
		suite.Require().NoError(err)
	}
}

// NewCoordinator initializes Coordinator with interchain security dummy provider and 2 neutron consumer chains
func NewProviderConsumerCoordinator(t *testing.T) *ibctesting.Coordinator {
	coordinator := ibctesting.NewCoordinator(t, 3)
	chainID := ibctesting.GetChainID(1)
	coordinator.Chains[chainID] = NewTestChain(t, coordinator, icssimapp.ProviderAppIniter, chainID)
	providerChain := coordinator.GetChain(chainID)

	chainID = ibctesting.GetChainID(2)
	coordinator.Chains[chainID] = NewTestChainWithValSet(t, coordinator,
		SetupTestingApp, chainID, providerChain.Vals, providerChain.Signers)

	chainID = ibctesting.GetChainID(3)
	coordinator.Chains[chainID] = NewTestChainWithValSet(t, coordinator,
		SetupTestingApp, chainID, providerChain.Vals, providerChain.Signers)

	return coordinator
}

func (suite *IBCConnectionTestSuite) GetNeutronZoneApp(chain *ibctesting.TestChain) *app.App {
	testApp, ok := chain.App.(*app.App)
	if !ok {
		panic("not NeutronZone app")
	}

	return testApp
}

func (suite *IBCConnectionTestSuite) StoreReflectCode(ctx sdk.Context, addr sdk.AccAddress, path string) uint64 {
	// wasm file built with https://github.com/neutron-org/neutron-contracts/tree/main/contracts/reflect
	wasmCode, err := os.ReadFile(path)
	suite.Require().NoError(err)

	codeID, _, err := keeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper).Create(ctx, addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody, Address: ""})
	suite.Require().NoError(err)

	return codeID
}

func (suite *IBCConnectionTestSuite) InstantiateReflectContract(ctx sdk.Context, funder sdk.AccAddress, codeID uint64) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	suite.Require().NoError(err)

	return addr
}

func NewICAPath(chainA, chainB, chainProvider *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.PortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	trustingPeriodFraction := chainProvider.App.(*appProvider.App).GetProviderKeeper().GetTrustingPeriodFraction(chainProvider.GetContext())

	consumerUnbondingPeriodA := path.EndpointA.Chain.App.(*app.App).GetConsumerKeeper().GetUnbondingPeriod(path.EndpointA.Chain.GetContext())
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriodA
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccv.CalculateTrustPeriod(consumerUnbondingPeriodA, trustingPeriodFraction)

	consumerUnbondingPeriodB := path.EndpointB.Chain.App.(*app.App).GetConsumerKeeper().GetUnbondingPeriod(path.EndpointB.Chain.GetContext())
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriodB
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccv.CalculateTrustPeriod(consumerUnbondingPeriodB, trustingPeriodFraction)

	return path
}

// SetupICAPath invokes the InterchainAccounts entrypoint and subsequent channel handshake handlers
func SetupICAPath(path *ibctesting.Path, owner string) error {
	if err := RegisterInterchainAccount(path.EndpointA, owner); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	return path.EndpointB.ChanOpenConfirm()
}

// RegisterInterchainAccount is a helper function for starting the channel handshake
func RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) error {
	icaOwner, _ := ictxstypes.NewICAOwner(owner, TestInterchainID)
	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		return err
	}

	ctx := endpoint.Chain.GetContext()

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(ctx)

	a, ok := endpoint.Chain.App.(*app.App)
	if !ok {
		return fmt.Errorf("not NeutronZoneApp")
	}

	// TODO(pr0n00gler): are we sure it's okay?
	if err := a.ICAControllerKeeper.RegisterInterchainAccount(ctx, endpoint.ConnectionID, icaOwner.String(), ""); err != nil {
		return err
	}

	// commit state changes for proof verification
	endpoint.Chain.NextBlock()

	// update port/channel ids
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID

	return nil
}

// SetupTestingApp initializes the IBC-go testing application
func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := app.MakeEncodingConfig()
	db := dbm.NewMemDB()
	testApp := app.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		app.GetEnabledProposals(),
		simapp.EmptyAppOptions{},
		nil,
	)
	return testApp, app.NewDefaultGenesisState(testApp.AppCodec())
}

func NewTransferPath(chainA, chainB, chainProvider *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = types.PortID
	path.EndpointB.ChannelConfig.PortID = types.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	trustingPeriodFraction := chainProvider.App.(*appProvider.App).GetProviderKeeper().GetTrustingPeriodFraction(chainProvider.GetContext())
	consumerUnbondingPeriodA := path.EndpointA.Chain.App.(*app.App).GetConsumerKeeper().GetUnbondingPeriod(path.EndpointA.Chain.GetContext())
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriodA
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccv.CalculateTrustPeriod(consumerUnbondingPeriodA, trustingPeriodFraction)

	consumerUnbondingPeriodB := path.EndpointB.Chain.App.(*app.App).GetConsumerKeeper().GetUnbondingPeriod(path.EndpointB.Chain.GetContext())
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriodB
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccv.CalculateTrustPeriod(consumerUnbondingPeriodB, trustingPeriodFraction)

	return path
}

// SetupTransferPath
func SetupTransferPath(path *ibctesting.Path) error {
	channelSequence := path.EndpointA.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(path.EndpointA.Chain.GetContext())
	channelSequenceB := path.EndpointB.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(path.EndpointB.Chain.GetContext())

	// update port/channel ids
	path.EndpointA.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	path.EndpointB.ChannelID = channeltypes.FormatChannelIdentifier(channelSequenceB)

	if err := path.EndpointA.ChanOpenInit(); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	return path.EndpointB.ChanOpenConfirm()
}
