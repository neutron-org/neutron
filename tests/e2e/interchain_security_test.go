package e2e_test

import (
	"encoding/json"
	"testing"

	"github.com/CosmWasm/wasmd/app"
	"github.com/cosmos/cosmos-sdk/simapp"
	appProvider "github.com/cosmos/interchain-security/app/provider"
	"github.com/cosmos/interchain-security/tests/e2e"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	appConsumer "github.com/neutron-org/neutron/app"

	ibctesting "github.com/cosmos/interchain-security/legacy_ibc_testing/testing"
	icssimapp "github.com/cosmos/interchain-security/testutil/ibc_testing"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {

	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
	ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
		// Pass in ibctesting.AppIniters for provider and consumer.
		icssimapp.ProviderAppIniter, ConsumerAppIniter, []string{})

	// Run tests
	suite.Run(t, ccvSuite)
}

func ConsumerAppIniter() (ibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := appConsumer.MakeEncodingConfig()
	testApp := appConsumer.New(
		log.NewNopLogger(),
		tmdb.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		5,
		encoding,
		app.GetEnabledProposals(),
		simapp.EmptyAppOptions{},
		nil,
	)
	return testApp, appConsumer.NewDefaultGenesisState(encoding.Marshaler)
}
