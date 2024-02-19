package app_test

import (
	"testing"

	cmttypes "github.com/cometbft/cometbft/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	icssimapp "github.com/cosmos/interchain-security/v4/testutil/ibc_testing"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v2/app"
	"github.com/neutron-org/neutron/v2/testutil"
)

func TestConsumerWhitelistingKeys(t *testing.T) {
	_ = app.GetDefaultConfig()
	coordinator := ibctesting.NewCoordinator(t, 2)
	chainID := ibctesting.GetChainID(1)

	ibctesting.DefaultTestingAppInit = icssimapp.ProviderAppIniter
	coordinator.Chains[chainID] = ibctesting.NewTestChain(t, coordinator, chainID)
	providerChain := coordinator.GetChain(chainID)

	ibctesting.DefaultTestingAppInit = testutil.SetupTestingApp(cmttypes.TM2PB.ValidatorUpdates(providerChain.Vals))
	chain := ibctesting.NewTestChain(t, coordinator, "test")

	paramKeeper := chain.App.(*app.App).ParamsKeeper
	for paramKey := range app.WhitelistedParams {
		ss, ok := paramKeeper.GetSubspace(paramKey.Subspace)
		require.True(t, ok, "Unknown subspace %s", paramKey.Subspace)
		hasKey := ss.Has(chain.GetContext(), []byte(paramKey.Key))
		require.True(t, hasKey, "Invalid key %s for subspace %s", paramKey.Key, paramKey.Subspace)
	}
}
