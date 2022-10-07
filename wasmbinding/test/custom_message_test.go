package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmvm/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/wasmbinding"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	ictxkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
)

type CustomMessengerTestSuite struct {
	testutil.IBCConnectionTestSuite
	neutron         *app.App
	ctx             sdk.Context
	messenger       *wasmbinding.CustomMessenger
	contractOwner   sdk.AccAddress
	contractAddress sdk.AccAddress
}

func (suite *CustomMessengerTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
	suite.neutron = suite.GetNeutronZoneApp(suite.ChainA)
	suite.ctx = suite.ChainA.GetContext()
	suite.messenger = &wasmbinding.CustomMessenger{}
	suite.messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(suite.neutron.InterchainTxsKeeper)
	suite.messenger.Keeper = suite.neutron.InterchainTxsKeeper
	suite.messenger.Icqmsgserver = icqkeeper.NewMsgServerImpl(suite.neutron.InterchainQueriesKeeper)
	suite.contractOwner = keeper.RandomAccountAddress(suite.T())
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainAccount() {
	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateReflectContract(suite.ctx, suite.contractOwner, codeId)
	suite.Require().NotEmpty(suite.contractAddress)

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
		testutil.TestInterchainID,
	))
	var msg json.RawMessage
	err := json.Unmarshal(msgStr, &msg)
	suite.NoError(err)

	// Dispatch RegisterInterchainAccount message
	events, data, err := suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{}`)}, data)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainQuery() {
	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateReflectContract(suite.ctx, suite.contractOwner, codeId)
	suite.Require().NotEmpty(suite.contractAddress)

	err := testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	// Top up contract balance
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(10_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)

	// Craft RegisterInterchainQuery message
	clientKey := host.FullClientStateKey(suite.Path.EndpointB.ClientID)
	updatePeriod := uint64(20)

	regMsg := bindings.RegisterInterchainQuery{
		QueryType: string(icqtypes.InterchainQueryTypeKV),
		Keys: []*icqtypes.KVKey{
			{Path: host.StoreKey, Key: clientKey},
		},
		TransactionsFilter: "{}",
		ConnectionId:       suite.Path.EndpointA.ConnectionID,
		UpdatePeriod:       updatePeriod,
	}

	fullMsg := bindings.NeutronMsg{
		RegisterInterchainQuery: &regMsg,
	}

	msg, err := json.Marshal(fullMsg)
	suite.NoError(err)

	events, data, err := suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{"id":1}`)}, data)
}

func (suite *CustomMessengerTestSuite) TestUpdateInterchainQuery() {
	// reuse register interchain query test to get query registered
	suite.TestRegisterInterchainQuery()
	// Craft UpdateInterchainQuery message
	queryID := uint64(1)
	newUpdatePeriod := uint64(111)
	updMsg := bindings.UpdateInterchainQuery{
		QueryId:         queryID,
		NewKeys:         nil,
		NewUpdatePeriod: newUpdatePeriod,
	}

	fullMsg := bindings.NeutronMsg{
		UpdateInterchainQuery: &updMsg,
	}

	msg, err := json.Marshal(fullMsg)
	suite.NoError(err)

	// Dispatch UpdateInterchainQuery message
	events, data, err := suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{}`)}, data)
}

func (suite *CustomMessengerTestSuite) TestUpdateInterchainQueryFailed() {
	// Craft UpdateInterchainQuery message
	queryID := uint64(1)
	newUpdatePeriod := uint64(111)
	updMsg := bindings.UpdateInterchainQuery{
		QueryId:         queryID,
		NewKeys:         nil,
		NewUpdatePeriod: newUpdatePeriod,
	}

	fullMsg := bindings.NeutronMsg{
		UpdateInterchainQuery: &updMsg,
	}

	msg, err := json.Marshal(fullMsg)
	suite.NoError(err)

	// Dispatch UpdateInterchainQuery message
	owner, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	suite.NoError(err)
	events, data, err := suite.messenger.DispatchMsg(suite.ctx, owner, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	expectedErrMsg := "failed to update interchain query: failed to update interchain query: failed to get query by query id: there is no query with id: 1"
	suite.Require().ErrorContains(err, expectedErrMsg)
	suite.Nil(events)
	suite.Nil(data)
}

func (suite *CustomMessengerTestSuite) TestRemoveInterchainQuery() {
	// reuse register interchain query test to get query registered
	suite.TestRegisterInterchainQuery()
	// Craft RemoveInterchainQuery message
	queryID := uint64(1)
	remMsg := bindings.RemoveInterchainQuery{
		QueryId: queryID,
	}

	fullMsg := bindings.NeutronMsg{
		RemoveInterchainQuery: &remMsg,
	}

	msg, err := json.Marshal(fullMsg)
	suite.NoError(err)

	// Dispatch RemoveInterchainQuery message
	suite.NoError(err)
	events, data, err := suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal([][]byte{[]byte(`{}`)}, data)
}

func (suite *CustomMessengerTestSuite) TestRemoveInterchainQueryFailed() {
	// Craft RemoveInterchainQuery message
	queryID := uint64(1)
	remMsg := bindings.RemoveInterchainQuery{
		QueryId: queryID,
	}

	fullMsg := bindings.NeutronMsg{
		RemoveInterchainQuery: &remMsg,
	}

	msg, err := json.Marshal(fullMsg)
	suite.NoError(err)

	// Dispatch RemoveInterchainQuery message
	owner, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	suite.NoError(err)
	events, data, err := suite.messenger.DispatchMsg(suite.ctx, owner, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	expectedErrMsg := "failed to remove interchain query: failed to remove interchain query: failed to get query by query id: there is no query with id: 1"
	suite.Require().ErrorContains(err, expectedErrMsg)
	suite.Nil(events)
	suite.Nil(data)
}

func (suite *CustomMessengerTestSuite) TestSubmitTx() {
	// Store code and instantiate reflect contract
	codeId := suite.StoreReflectCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateReflectContract(suite.ctx, suite.contractOwner, codeId)
	suite.Require().NotEmpty(suite.contractAddress)

	err := testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	// Craft SubmitTx message
	memo := "Jimmy"
	timeout := 2000
	msgs := `[{"type_url":"/cosmos.staking.v1beta1.MsgDelegate","value":[26,10,10,5,115,116,97,107,101,18,1,48]}]`
	msgStr := []byte(fmt.Sprintf(
		`
{
	"submit_tx": {
		"connection_id": "%s",
		"interchain_account_id": "%s",
		"msgs": %s,
		"memo": "%s",
		"timeout": %d
	}
}
		`,
		suite.Path.EndpointA.ConnectionID,
		testutil.TestInterchainID,
		msgs,
		memo,
		timeout,
	))
	var msg json.RawMessage
	err = json.Unmarshal(msgStr, &msg)
	suite.NoError(err)

	// Dispatch SubmitTx message
	events, data, err := suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.NoError(err)

	var response bindings.SubmitTxResponse
	err = json.Unmarshal(data[0], &response)
	suite.NoError(err)

	suite.NoError(err)
	suite.Nil(events)
	suite.Equal(uint64(1), response.SequenceId)
	suite.Equal("channel-0", response.Channel)
}

func TestMessengerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomMessengerTestSuite))
}
