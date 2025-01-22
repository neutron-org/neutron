package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v5/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

func TestSubscribedContractsQuery(t *testing.T) {
	keeper, ctx := keepertest.HarpoonKeeper(t, nil, nil)

	// before set return empty
	response, err := keeper.SubscribedContracts(ctx, &types.QuerySubscribedContractsRequest{HookType: types.HookType_AfterValidatorCreated.String()})
	require.NoError(t, err)
	require.Equal(t, &types.QuerySubscribedContractsResponse{ContractAddresses: []string{}}, response)

	// add hook
	keeper.UpdateHookSubscription(ctx, &types.HookSubscription{
		ContractAddress: ContractAddress1,
		Hooks:           []types.HookType{types.HookType_AfterValidatorCreated},
	})

	// after adding returns hook
	response, err = keeper.SubscribedContracts(ctx, &types.QuerySubscribedContractsRequest{HookType: types.HookType_AfterValidatorCreated.String()})
	require.NoError(t, err)
	require.Equal(t, &types.QuerySubscribedContractsResponse{ContractAddresses: []string{ContractAddress1}}, response)
}
