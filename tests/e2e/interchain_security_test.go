package e2e_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appProvider "github.com/cosmos/interchain-security/v6/app/provider"
	e2e "github.com/cosmos/interchain-security/v6/tests/integration"
	icssimapp "github.com/cosmos/interchain-security/v6/testutil/ibc_testing"
	"github.com/stretchr/testify/suite"

	appConsumer "github.com/neutron-org/neutron/v4/app"
	appparams "github.com/neutron-org/neutron/v4/app/params"
	"github.com/neutron-org/neutron/v4/testutil"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {
	sdk.DefaultBondDenom = appparams.DefaultDenom
	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
	ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
		// Pass in ibctesting.AppIniters for provider and consumer.
		icssimapp.ProviderAppIniter, testutil.SetupValSetAppIniter,
		// TODO: These test just doesn't work in IS, so skip it for now
		// "TestKeyAssignment", "TestBasicSlashPacketThrottling", "TestMultiConsumerSlashPacketThrottling" - don't work because ICS doesn't support CometBFT v0.38.12 yet
		// "TestRewardsDistribution" - doesn't work because we burn some fees
		[]string{"TestRewardsDistribution", "TestKeyAssignment", "TestBasicSlashPacketThrottling", "TestMultiConsumerSlashPacketThrottling"},
	)

	// Run tests
	suite.Run(t, ccvSuite)
}
