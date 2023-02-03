package app_test

import (
	"encoding/json"
	"testing"

	simapp "github.com/cosmos/cosmos-sdk/simapp"
	ibctesting "github.com/cosmos/interchain-security/legacy_ibc_testing/testing"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/neutron-org/neutron/app"
)

func TestConsumerWhitelistingKeys(t *testing.T) {
	_ = app.GetDefaultConfig()
	chain := ibctesting.NewTestChain(t, ibctesting.NewCoordinator(t, 0), SetupTestingAppConsumer, "test")
	paramKeeper := chain.App.(*app.App).ParamsKeeper
	for paramKey := range app.WhitelistedParams {
		ss, ok := paramKeeper.GetSubspace(paramKey.Subspace)
		require.True(t, ok, "Unknown subspace %s", paramKey.Subspace)
		hasKey := ss.Has(chain.GetContext(), []byte(paramKey.Key))
		require.True(t, hasKey, "Invalid key %s for subspace %s", paramKey.Key, paramKey.Subspace)
	}
}

func SetupTestingAppConsumer() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := tmdb.NewMemDB()
	encoding := app.MakeEncodingConfig()
	testApp := app.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		app.GetEnabledProposals(),
		simapp.EmptyAppOptions{},
		nil,
	)
	return testApp, app.NewDefaultGenesisState(encoding.Marshaler)
}
