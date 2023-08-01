package e2e_test

import (
	"testing"

	appProvider "github.com/cosmos/interchain-security/v3/app/provider"
	icssimapp "github.com/cosmos/interchain-security/v3/testutil/ibc_testing"
	"github.com/stretchr/testify/suite"

	e2e "github.com/cosmos/interchain-security/v3/tests/integration"

	appConsumer "github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {
	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
	ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
		// Pass in ibctesting.AppIniters for provider and consumer.
		icssimapp.ProviderAppIniter, testutil.SetupTestingApp("test-1"),
		// TODO: These three tests just don't work in IS, so skip them for now
		[]string{"TestSendRewardsRetries", "TestRewardsDistribution", "TestEndBlockRD"})

	// Run tests
	suite.Run(t, ccvSuite)
}
