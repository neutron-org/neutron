package testutil

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	cometbfttypes "github.com/cometbft/cometbft/abci/types"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	types2 "github.com/cosmos/cosmos-sdk/crypto/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"

	"github.com/neutron-org/neutron/v6/utils"

	"github.com/neutron-org/neutron/v6/app/config"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	db2 "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
	"github.com/stretchr/testify/suite"

	appparams "github.com/neutron-org/neutron/v6/app/params"
	tokenfactorytypes "github.com/neutron-org/neutron/v6/x/tokenfactory/types"

	//nolint:staticcheck
	"github.com/neutron-org/neutron/v6/app"
	ictxstypes "github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

var (
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"

	TestInterchainID = "owner_id"

	ConnectionZero = "connection-0"

	// TestVersion defines a reusable interchainaccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ConnectionZero,
		HostConnectionId:       ConnectionZero,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

func init() {
	config.GetDefaultConfig()
	// Disable cache since enabled cache triggers test errors when `AccAddress.String()`
	// gets called before setting neutron bech32 prefix
	sdk.SetAddrCacheEnabled(false)
}

type IBCConnectionTestSuite struct {
	suite.Suite
	Coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	ChainProvider *ibctesting.TestChain
	ChainA        *ibctesting.TestChain
	ChainB        *ibctesting.TestChain
	ChainC        *ibctesting.TestChain

	Path         *ibctesting.Path
	TransferPath *ibctesting.Path

	TransferPathAC *ibctesting.Path
}

func (suite *IBCConnectionTestSuite) SetupTest() {
	// we need to redefine this variable to make tests work cause we use untrn as default bond denom in neutron
	sdk.DefaultBondDenom = appparams.DefaultDenom

	suite.Coordinator = NewProviderConsumerCoordinator(suite.T())
	suite.ChainProvider = suite.Coordinator.GetChain(ibctesting.GetChainID(1))
	suite.ChainA = suite.Coordinator.GetChain(ibctesting.GetChainID(2))
	suite.ChainB = suite.Coordinator.GetChain(ibctesting.GetChainID(3))
	suite.ChainC = suite.Coordinator.GetChain(ibctesting.GetChainID(4))

	suite.Path = NewICAPath(suite.ChainA, suite.ChainB)

	suite.Coordinator.SetupConnections(suite.Path)
}

func (suite *IBCConnectionTestSuite) ConfigureTransferChannelAC() {
	suite.TransferPathAC = NewTransferPath(suite.ChainA, suite.ChainC)
	suite.Coordinator.SetupConnections(suite.TransferPathAC)
	err := SetupTransferPath(suite.TransferPathAC)
	suite.Require().NoError(err)
}

func (suite *IBCConnectionTestSuite) ConfigureTransferChannel() {
	suite.TransferPath = NewTransferPath(suite.ChainA, suite.ChainB)
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

// FundModuleAcc funds target modules with specified amount.
func (suite *IBCConnectionTestSuite) FundModuleAcc(moduleName string, amounts sdk.Coins) {
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.MintCoins(suite.ChainA.GetContext(), tokenfactorytypes.ModuleName, amounts)
	suite.Require().NoError(err)
	err = bankKeeper.SendCoinsFromModuleToModule(suite.ChainA.GetContext(), tokenfactorytypes.ModuleName, moduleName, amounts)
	suite.Require().NoError(err)
}

func testHomeDir(chainID string) string {
	projectRoot := utils.RootDir()
	return path.Join(projectRoot, ".testchains", chainID)
}

// NewCoordinator initializes Coordinator with interchain security dummy provider and 2 neutron consumer chains
func NewProviderConsumerCoordinator(t *testing.T) *ibctesting.Coordinator {
	coordinator := ibctesting.NewCoordinator(t, 0)
	chainID := ibctesting.GetChainID(1)

	ibctesting.DefaultTestingAppInit = SetupTestingApp()
	coordinator.Chains[chainID] = ibctesting.NewTestChain(t, coordinator, chainID)
	providerChain := coordinator.GetChain(chainID)

	_ = config.GetDefaultConfig()
	sdk.SetAddrCacheEnabled(false)
	chainID = ibctesting.GetChainID(2)
	coordinator.Chains[chainID] = ibctesting.NewTestChainWithValSet(t, coordinator,
		chainID, providerChain.Vals, providerChain.Signers)

	chainID = ibctesting.GetChainID(3)
	coordinator.Chains[chainID] = ibctesting.NewTestChainWithValSet(t, coordinator,
		chainID, providerChain.Vals, providerChain.Signers)

	chainID = ibctesting.GetChainID(4)
	coordinator.Chains[chainID] = ibctesting.NewTestChainWithValSet(t, coordinator,
		chainID, providerChain.Vals, providerChain.Signers)

	return coordinator
}

func (suite *IBCConnectionTestSuite) GetNeutronZoneApp(chain *ibctesting.TestChain) *app.App {
	testApp, ok := chain.App.(*app.App)
	if !ok {
		panic("not NeutronZone app")
	}

	return testApp
}

func (suite *IBCConnectionTestSuite) StoreTestCode(ctx sdk.Context, addr sdk.AccAddress, path string) uint64 {
	// wasm file built with https://github.com/neutron-org/neutron-sdk/tree/main/contracts/reflect
	// wasm file built with https://github.com/neutron-org/neutron-dev-contracts/tree/feat/ica-register-fee-update/contracts/neutron_interchain_txs
	wasmCode, err := os.ReadFile(path)
	suite.Require().NoError(err)

	codeID, _, err := keeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper).Create(ctx, addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody})
	suite.Require().NoError(err)

	return codeID
}

func (suite *IBCConnectionTestSuite) InstantiateTestContract(ctx sdk.Context, funder sdk.AccAddress, codeID uint64) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	suite.Require().NoError(err)

	return addr
}

func NewICAPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	// trustingPeriodFraction := chainProvider.App.(*app.App).GetStakingKeeper().(stakingkeeper.Keeper).GetUnbonding(chainProvider.GetContext())
	trustingPeriodFraction := 0.66
	paramsA, err := path.EndpointA.Chain.App.(*app.App).GetStakingKeeper().GetParams(path.EndpointA.Chain.GetContext())
	if err != nil {
		panic(err)
	}
	UnbondingPeriodA := paramsA.UnbondingTime
	trustingA := time.Duration(float64(UnbondingPeriodA.Nanoseconds()) * trustingPeriodFraction)
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = UnbondingPeriodA
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod = trustingA

	paramsB, err := path.EndpointB.Chain.App.(*app.App).GetStakingKeeper().GetParams(path.EndpointB.Chain.GetContext())
	if err != nil {
		panic(err)
	}
	UnbondingPeriodB := paramsB.UnbondingTime
	trustingB := time.Duration(float64(UnbondingPeriodB.Nanoseconds()) * trustingPeriodFraction)
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = UnbondingPeriodB
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod = trustingB

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

	icaMsgServer := icacontrollerkeeper.NewMsgServerImpl(&a.ICAControllerKeeper)
	if _, err = icaMsgServer.RegisterInterchainAccount(ctx, &icacontrollertypes.MsgRegisterInterchainAccount{
		Owner:        icaOwner.String(),
		ConnectionId: endpoint.ConnectionID,
		Version:      TestVersion,
		Ordering:     channeltypes.ORDERED,
	}); err != nil {
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
func SetupTestingApp() func() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		sdk.DefaultBondDenom = appparams.DefaultDenom
		encoding := app.MakeEncodingConfig()
		db := db2.NewMemDB()
		homePath := testHomeDir("testchain-" + tmrand.NewRand().Str(6))
		testApp := app.New(
			log.NewNopLogger(),
			db,
			nil,
			false,
			map[int64]bool{},
			homePath,
			0,
			encoding,
			sims.EmptyAppOptions{},
			nil,
		)

		// we need to set up a TestInitChainer where we can redefine MaxBlockGas in ConsensusParamsKeeper
		testApp.SetInitChainer(testApp.TestInitChainer)
		// and then we manually init baseapp and load states
		testApp.LoadLatest()

		genesisState := app.NewDefaultGenesisState(testApp.AppCodec())

		return testApp, genesisState
	}
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = types.PortID
	path.EndpointB.ChannelConfig.PortID = types.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = types.V1
	path.EndpointB.ChannelConfig.Version = types.V1

	trustingPeriodFraction := 0.66
	paramsA, err := path.EndpointA.Chain.App.(*app.App).GetStakingKeeper().GetParams(path.EndpointA.Chain.GetContext())
	if err != nil {
		panic(err)
	}
	UnbondingPeriodA := paramsA.UnbondingTime
	trustingA := time.Duration(float64(UnbondingPeriodA.Nanoseconds()) * trustingPeriodFraction)
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = UnbondingPeriodA
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod = trustingA

	paramsB, err := path.EndpointB.Chain.App.(*app.App).GetStakingKeeper().GetParams(path.EndpointB.Chain.GetContext())
	if err != nil {
		panic(err)
	}
	UnbondingPeriodB := paramsB.UnbondingTime
	trustingB := time.Duration(float64(UnbondingPeriodB.Nanoseconds()) * trustingPeriodFraction)
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = UnbondingPeriodB
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod = trustingB

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

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (suite *IBCConnectionTestSuite) SendMsgsNoCheck(chain *ibctesting.TestChain, msgs ...sdk.Msg) (*cometbfttypes.ExecTxResult, error) {
	// ensure the suite has the latest time
	suite.Coordinator.UpdateTimeForChain(chain)

	// increment acc sequence regardless of success or failure tx execution
	defer func() {
		err := chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := SignAndDeliver(chain.TB, chain.TxConfig, chain.App.GetBaseApp(), msgs, chain.ChainID, []uint64{chain.SenderAccount.GetAccountNumber()}, []uint64{chain.SenderAccount.GetSequence()}, chain.LatestCommittedHeader.GetTime(), chain.NextVals.Hash(), chain.SenderPrivKey)
	if err != nil {
		return nil, err
	}

	suite.commitBlock(resp, chain)

	suite.Coordinator.IncrementTime()

	suite.Require().Len(resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	suite.Coordinator.IncrementTime()

	return txResult, nil
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliver(
	tb testing.TB,
	txCfg client.TxConfig,
	app *baseapp.BaseApp,
	msgs []sdk.Msg,
	chainID string,
	accNums, accSeqs []uint64,
	blockTime time.Time,
	nextValHash []byte,
	priv ...types2.PrivKey,
) (res *cometbfttypes.ResponseFinalizeBlock, err error) {
	tb.Helper()
	tx, err := sims.GenSignedMockTx(
		// #nosec G404 - math/rand is acceptable for non-cryptographic purposes
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		sims.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	if err != nil {
		return nil, err
	}

	txBytes, err := txCfg.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	return app.FinalizeBlock(&cometbfttypes.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               blockTime,
		NextValidatorsHash: nextValHash,
		Txs:                [][]byte{txBytes},
	})
}

func (suite *IBCConnectionTestSuite) ExecuteContract(contract, sender sdk.AccAddress, msg []byte, funds sdk.Coins) ([]byte, error) {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	return contractKeeper.Execute(suite.ChainA.GetContext(), contract, sender, msg, funds)
}

func (suite *IBCConnectionTestSuite) commitBlock(res *cometbfttypes.ResponseFinalizeBlock, chain *ibctesting.TestChain) {
	_, err := chain.App.Commit()
	require.NoError(chain.TB, err)

	// set the last header to the current header
	// use nil trusted fields
	chain.LatestCommittedHeader = chain.CurrentTMClientHeader()
	// set the trusted validator set to the next validator set
	// The latest trusted validator set is the next validator set
	// associated with the header being committed in storage. This will
	// allow for header updates to be proved against these validators.
	chain.TrustedValidators[uint64(chain.ProposedHeader.Height)] = chain.NextVals

	// val set changes returned from previous block get applied to the next validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain, chain.Vals, res.ValidatorUpdates)

	// increment the proposer priority of validators
	chain.Vals.IncrementProposerPriority(1)

	// increment the current header
	chain.ProposedHeader = cmtproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.ProposedHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.Vals.Proposer.Address,
	}
}
