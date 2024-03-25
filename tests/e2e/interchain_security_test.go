package e2e_test

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
//TODO: func TestCCVTestSuite(t *testing.T) {
//	sdk.DefaultBondDenom = appparams.DefaultDenom
//	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
//	ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
//		// Pass in ibctesting.AppIniters for provider and consumer.
//		icssimapp.ProviderAppIniter, testutil.SetupValSetAppIniter,
//		// TODO: These three tests just don't work in IS, so skip them for now
//		[]string{"TestSendRewardsRetries", "TestRewardsDistribution", "TestEndBlockRD"})
//
//	// Run tests
//	suite.Run(t, ccvSuite)
//}
