package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v6/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

func TestSubscribedContractsQuery(t *testing.T) {
	k, ctx := keepertest.HarpoonKeeper(t, nil)
	queryServer := keeper.NewQueryServerImpl(k)

	// before set return empty
	response, err := queryServer.SubscribedContracts(ctx, &types.QuerySubscribedContractsRequest{HookType: types.HOOK_TYPE_AFTER_VALIDATOR_CREATED})
	require.NoError(t, err)
	require.Equal(t, &types.QuerySubscribedContractsResponse{ContractAddresses: []string{}}, response)

	// add hook
	k.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED},
	})

	// after adding returns hook
	response, err = queryServer.SubscribedContracts(ctx, &types.QuerySubscribedContractsRequest{HookType: types.HOOK_TYPE_AFTER_VALIDATOR_CREATED})
	require.NoError(t, err)
	require.Equal(t, &types.QuerySubscribedContractsResponse{ContractAddresses: []string{ContractAddress1}}, response)
}

func TestSubscribedContractsQueryUnspecifiedHookTypeFails(t *testing.T) {
	k, ctx := keepertest.HarpoonKeeper(t, nil)
	queryServer := keeper.NewQueryServerImpl(k)

	response, err := queryServer.SubscribedContracts(ctx, &types.QuerySubscribedContractsRequest{HookType: types.HOOK_TYPE_UNSPECIFIED})
	require.Error(t, err)
	require.Nil(t, response)
}

func TestSubscribedContractsQueryInvalidHookTypeFails(t *testing.T) {
	k, ctx := keepertest.HarpoonKeeper(t, nil)
	queryServer := keeper.NewQueryServerImpl(k)

	response, err := queryServer.SubscribedContracts(ctx, &types.QuerySubscribedContractsRequest{HookType: -200})
	require.Error(t, err)
	require.Nil(t, response)
}
