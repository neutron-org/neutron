package test

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	ictxkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	config := app.GetDefaultConfig()
	config.Seal()
}

func TestRegisterInterchainAccount(t *testing.T) {
	// Setup IBC chains and create connection between them
	ibcStruct := testutil.SetupIBCConnection(t)
	neutron, ok := ibcStruct.ChainA.App.(*app.App)
	require.True(t, ok)

	// Craft RegisterInterchainAccount message
	msgStr := []byte(fmt.Sprintf(
		`
{
	"register_interchain_account": {
		"connection_id": "%s",
		"interchain_account_id": "%s"
	}
}
		`,
		ibcStruct.Path.EndpointA.ConnectionID,
		testutil.TestInterchainId,
	))
	var msg json.RawMessage
	err := json.Unmarshal(msgStr, &msg)
	require.NoError(t, err)

	// Dispatch RegisterInterchainAccount message
	ctx := neutron.NewContext(true, ibcStruct.ChainA.CurrentHeader)
	messenger := wasmbinding.CustomMessenger{}
	messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(neutron.InterchainTxsKeeper)
	events, data, err := messenger.DispatchMsg(ctx, keeper.RandomAccountAddress(t), ibcStruct.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	require.NoError(t, err)
	require.Nil(t, events)
	require.Equal(t, [][]byte{[]byte(`{}`)}, data)
}

func TestRegisterInterchainQuery(t *testing.T) {
	// Setup IBC chains and create connection between them
	ibcStruct := testutil.SetupIBCConnection(t)
	neutron, ok := ibcStruct.ChainA.App.(*app.App)
	require.True(t, ok)

	// Craft RegisterInterchainQuery message
	queryType := "/cosmos.staking.v1beta1.Query/AllDelegations"
	queryData := "{}"
	updatePeriod := 20
	msgStr := []byte(fmt.Sprintf(
		`
{
	"register_interchain_query": {
		"query_type": "%s",
		"query_data": "%s",
		"zone_id": "%s",
		"connection_id": "%s",
		"update_period": %d
	}
}
		`,
		queryType,
		queryData,
		ibcStruct.ChainB.ChainID,
		ibcStruct.Path.EndpointA.ConnectionID,
		updatePeriod,
	))
	var msg json.RawMessage
	err := json.Unmarshal(msgStr, &msg)
	require.NoError(t, err)

	// Dispatch RegisterInterchainQuery message
	owner, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	require.NoError(t, err)
	ctx := neutron.NewContext(true, ibcStruct.ChainA.CurrentHeader)
	messenger := wasmbinding.CustomMessenger{}
	messenger.Icqmsgserver = icqkeeper.NewMsgServerImpl(neutron.InterchainQueriesKeeper)
	events, data, err := messenger.DispatchMsg(ctx, owner, ibcStruct.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	require.NoError(t, err)
	require.Nil(t, events)
	require.Equal(t, [][]byte{[]byte(`{"id":1}`)}, data)
}

func TestSubmitTx(t *testing.T) {
	// Setup IBC chains and create connection between them
	ibcStruct := testutil.SetupIBCConnection(t)
	neutron, ok := ibcStruct.ChainA.App.(*app.App)
	require.True(t, ok)

	// Craft SubmitTx message
	memo := "Jimmy"
	msgs := `[{"type_url":"/cosmos.staking.v1beta1.MsgDelegate","value":[26,10,10,5,115,116,97,107,101,18,1,48]}]`
	msgStr := []byte(fmt.Sprintf(
		`
{
	"submit_tx": {
		"connection_id": "%s",
		"interchain_account_id": "%s",
		"msgs": %s,
		"memo": "%s"
	}
}
		`,
		ibcStruct.Path.EndpointA.ConnectionID,
		testutil.TestInterchainId,
		msgs,
		memo,
	))
	var msg json.RawMessage
	err := json.Unmarshal(msgStr, &msg)
	require.NoError(t, err)

	// Dispatch SubmitTx message
	owner, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	require.NoError(t, err)
	ctx := neutron.NewContext(true, ibcStruct.ChainA.CurrentHeader)
	messenger := wasmbinding.CustomMessenger{}
	messenger.Keeper = neutron.InterchainTxsKeeper
	messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(neutron.InterchainTxsKeeper)
	messenger.Icqmsgserver = icqkeeper.NewMsgServerImpl(neutron.InterchainQueriesKeeper)
	events, data, err := messenger.DispatchMsg(ctx, owner, ibcStruct.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	require.NoError(t, err)
	require.Nil(t, events)
	require.Equal(t, [][]byte{[]byte(`{}`)}, data)
}
