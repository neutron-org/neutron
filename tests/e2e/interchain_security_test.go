package e2e_test

import (
	"testing"

	appProvider "github.com/cosmos/interchain-security/app/provider"
	"github.com/cosmos/interchain-security/tests/e2e"
	"github.com/stretchr/testify/suite"

	appConsumer "github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"

	icssimapp "github.com/cosmos/interchain-security/testutil/ibc_testing"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {
	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
	ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
		// Pass in ibctesting.AppIniters for provider and consumer.
		icssimapp.ProviderAppIniter, testutil.SetupTestingApp,
		// TODO: These three tests just don't work in IS, so skip them for now
		[]string{"TestSendRewardsRetries", "TestRewardsDistribution", "TestEndBlockRD"})

	// Run tests
	suite.Run(t, ccvSuite)
}
