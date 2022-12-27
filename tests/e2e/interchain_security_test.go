package e2e_test

import (
	"encoding/json"
	"testing"

	cosmossimapp "github.com/cosmos/cosmos-sdk/simapp"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	appProvider "github.com/cosmos/interchain-security/app/provider"
	"github.com/cosmos/interchain-security/tests/e2e"
	e2etestutil "github.com/cosmos/interchain-security/testutil/e2e"
	"github.com/cosmos/interchain-security/testutil/simapp"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/neutron-org/neutron/app"
	appConsumer "github.com/neutron-org/neutron/app"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {
	ccvSuite := e2e.NewCCVTestSuite(
		func(t *testing.T) (
			*ibctesting.Coordinator,
			*ibctesting.TestChain,
			*ibctesting.TestChain,
			e2etestutil.ProviderApp,
			e2etestutil.ConsumerApp,
		) {
			// Here we pass the concrete types that must implement the necessary interfaces
			// to be ran with e2e tests.
			coord, prov, cons := NewProviderConsumerCoordinator(t)
			return coord, prov, cons, prov.App.(*appProvider.App), cons.App.(*appConsumer.App)
		},
	)
	suite.Run(t, ccvSuite)
}

// NewCoordinator initializes Coordinator with interchain security dummy provider and neutron consumer chain
func NewProviderConsumerCoordinator(t *testing.T) (*ibctesting.Coordinator, *ibctesting.TestChain, *ibctesting.TestChain) {
	coordinator := simapp.NewBasicCoordinator(t)
	chainID := ibctesting.GetChainID(1)
	coordinator.Chains[chainID] = ibctesting.NewTestChain(t, coordinator, simapp.SetupTestingappProvider, chainID)
	providerChain := coordinator.GetChain(chainID)
	chainID = ibctesting.GetChainID(2)
	coordinator.Chains[chainID] = ibctesting.NewTestChainWithValSet(t, coordinator,
		SetupTestingAppConsumer, chainID, providerChain.Vals, providerChain.Signers)
	consumerChain := coordinator.GetChain(chainID)
	return coordinator, providerChain, consumerChain
}

func SetupTestingAppConsumer() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encCdc := appConsumer.MakeEncodingConfig()
	app := appConsumer.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		appConsumer.DefaultNodeHome,
		0,
		encCdc,
		app.GetEnabledProposals(),
		cosmossimapp.EmptyAppOptions{},
		nil)

	return app, appConsumer.NewDefaultGenesisState(encCdc.Marshaler)
}
