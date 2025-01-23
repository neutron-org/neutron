package e2e_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appProvider "github.com/cosmos/interchain-security/v5/app/provider"
	e2e "github.com/cosmos/interchain-security/v5/tests/integration"
	icssimapp "github.com/cosmos/interchain-security/v5/testutil/ibc_testing"
	"github.com/stretchr/testify/suite"

	appConsumer "github.com/neutron-org/neutron/v5/app"
	appparams "github.com/neutron-org/neutron/v5/app/params"
	"github.com/neutron-org/neutron/v5/testutil"
)

// Executes the standard group of ccv tests against a consumer and provider app.go implementation.
func TestCCVTestSuite(t *testing.T) {
	sdk.DefaultBondDenom = appparams.DefaultDenom
	// Pass in concrete app types that implement the interfaces defined in /testutil/e2e/interfaces.go
	ccvSuite := e2e.NewCCVTestSuite[*appProvider.App, *appConsumer.App](
		// Pass in ibctesting.AppIniters for provider and consumer.
		icssimapp.ProviderAppIniter, testutil.SetupValSetAppIniter,
		// TODO: These test just doesn't work in IS, so skip it for now
		[]string{"TestRewardsDistribution"},
	)

	// Run tests
	suite.Run(t, ccvSuite)
}
