package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	ictxkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

type CustomMessengerTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainAccount() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)
	messenger := wasmbinding.CustomMessenger{}
	messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(neutron.InterchainTxsKeeper)

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
		suite.Path.EndpointA.ConnectionID,
		testutil.TestInterchainId,
	))
	var msg json.RawMessage
	err := json.Unmarshal(msgStr, &msg)
	suite.NoError(err)

	// Dispatch RegisterInterchainAccount message
	events, data, err := messenger.DispatchMsg(ctx, keeper.RandomAccountAddress(suite.T()), suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{}`)}, data)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainQuery() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)
	messenger := wasmbinding.CustomMessenger{}
	messenger.Icqmsgserver = icqkeeper.NewMsgServerImpl(neutron.InterchainQueriesKeeper)

	// Craft RegisterInterchainQuery message
	clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
	updatePeriod := uint64(20)

	regMsg := bindings.RegisterInterchainQuery{
		QueryType: icqtypes.InterchainQueryTypeKV,
		Keys: []*icqtypes.KVKey{
			{Path: host.StoreKey, Key: clientKey},
		},
		TransactionsFilter: "{}",
		ZoneId:             suite.ChainB.ChainID,
		ConnectionId:       suite.Path.EndpointA.ConnectionID,
		UpdatePeriod:       updatePeriod,
	}

	fullMsg := bindings.NeutronMsg{
		RegisterInterchainQuery: &regMsg,
	}

	msg, err := json.Marshal(fullMsg)
	suite.NoError(err)

	// Dispatch RegisterInterchainQuery message
	owner, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	suite.NoError(err)
	events, data, err := messenger.DispatchMsg(ctx, owner, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{"id":1}`)}, data)
}

func (suite *CustomMessengerTestSuite) TestSubmitTx() {
	neutron := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := neutron.NewContext(true, suite.ChainA.CurrentHeader)

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
		suite.Path.EndpointA.ConnectionID,
		testutil.TestInterchainId,
		msgs,
		memo,
	))
	var msg json.RawMessage
	err := json.Unmarshal(msgStr, &msg)
	suite.NoError(err)

	// Dispatch SubmitTx message
	owner, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	suite.NoError(err)
	messenger := wasmbinding.CustomMessenger{}
	messenger.Keeper = neutron.InterchainTxsKeeper
	messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(neutron.InterchainTxsKeeper)
	messenger.Icqmsgserver = icqkeeper.NewMsgServerImpl(neutron.InterchainQueriesKeeper)
	events, data, err := messenger.DispatchMsg(ctx, owner, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{}`)}, data)
}

func TestMessengerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomMessengerTestSuite))
}
