package interchainqueries_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	keepertest "github.com/neutron-org/neutron/v6/testutil/interchainqueries/keeper"
	"github.com/neutron-org/neutron/v6/x/interchainqueries"
	"github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		RegisteredQueries: []*types.RegisteredQuery{
			{
				Id: 4,
			},
			{
				Id: 3,
			},
			{
				Id: 2,
			},
			{
				Id: 1,
			},
		},
	}

	require.EqualValues(t, genesisState.Params, types.DefaultParams())

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	lastQueryID := k.GetLastRegisteredQueryKey(ctx)

	require.EqualValues(t, got.Params, types.DefaultParams())
	require.NotNil(t, got)
	require.EqualValues(t, 4, lastQueryID)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}

func TestGenesisNullQueries(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}

func TestGenesisFilledQueries(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		RegisteredQueries: []*types.RegisteredQuery{
			{
				Id:        4,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:        3,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:                 2,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
			{
				Id:                 1,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
		},
	}

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	err := got.Validate()
	require.NoError(t, err)

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}

func TestGenesisMalformedQueriesInvalidTxFilter(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		RegisteredQueries: []*types.RegisteredQuery{
			{
				Id:        4,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:        3,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:                 2,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
			{
				Id:                 1,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"fi><eld":"tx.height","op":"Eq","value":1000}]`,
			},
		},
	}

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	err := got.Validate()
	require.ErrorContains(t, err, "invalid transactions filter")

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}

func TestGenesisMalformedQueriesNoKvKeys(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		RegisteredQueries: []*types.RegisteredQuery{
			{
				Id:        4,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
			},
			{
				Id:        3,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:                 2,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
			{
				Id:                 1,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
		},
	}

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	err := got.Validate()
	require.ErrorContains(t, err, "keys are empty")

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}

func TestGenesisMalformedQueriesInvalidQueryType(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		RegisteredQueries: []*types.RegisteredQuery{
			{
				Id:        4,
				QueryType: "fake",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
			},
			{
				Id:        3,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:                 2,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
			{
				Id:                 1,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
		},
	}

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	err := got.Validate()
	require.ErrorContains(t, err, "Unexpected query type")

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}

func TestGenesisMalformedQueriesInvalidPrefix(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		RegisteredQueries: []*types.RegisteredQuery{
			{
				Id:        4,
				QueryType: "kv",
				Owner:     "neutron18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:        3,
				QueryType: "kv",
				Owner:     "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				Keys: []*types.KVKey{
					{
						Path: "newpath",
						Key:  []byte("newdata"),
					},
				},
			},
			{
				Id:                 2,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
			{
				Id:                 1,
				QueryType:          "tx",
				Owner:              "cosmos18g0avxazu3dkgd5n5ea8h8rtl78de0hytsj9vm",
				TransactionsFilter: `[{"field":"tx.height","op":"Eq","value":1000}]`,
			},
		},
	}

	k, ctx := keepertest.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	interchainqueries.InitGenesis(ctx, *k, genesisState)
	got := interchainqueries.ExportGenesis(ctx, *k)
	err := got.Validate()
	require.ErrorContains(t, err, "Invalid owner address")

	require.ElementsMatch(t, genesisState.RegisteredQueries, got.RegisteredQueries)
}
