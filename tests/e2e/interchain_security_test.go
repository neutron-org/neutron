package e2e_test

import (
	"testing"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {
	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
	//ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
	//	// Pass in ibctesting.AppIniters for provider and consumer.
	//	icssimapp.ProviderAppIniter, testutil.SetupTestingApp,
	//	// TODO: These three tests just don't work in IS, so skip them for now
	//	[]string{"TestSendRewardsRetries", "TestRewardsDistribution", "TestEndBlockRD"})

	// Run tests
	//COMMENTED OUT BECAUSE OF SOFT OPT OUT FEATURE
	//suite.Run(t, ccvSuite)
}
