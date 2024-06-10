package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	admintypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	keeper2 "github.com/neutron-org/neutron/v4/x/contractmanager/keeper"
	feeburnertypes "github.com/neutron-org/neutron/v4/x/feeburner/types"

	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/stretchr/testify/suite"

	ictxtypes "github.com/neutron-org/neutron/v4/x/interchaintxs/types"

	adminkeeper "github.com/cosmos/admin-module/v2/x/adminmodule/keeper"

	"github.com/neutron-org/neutron/v4/app/params"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/app"
	"github.com/neutron-org/neutron/v4/testutil"
	"github.com/neutron-org/neutron/v4/wasmbinding"
	"github.com/neutron-org/neutron/v4/wasmbinding/bindings"
	feetypes "github.com/neutron-org/neutron/v4/x/feerefunder/types"
	icqkeeper "github.com/neutron-org/neutron/v4/x/interchainqueries/keeper"
	icqtypes "github.com/neutron-org/neutron/v4/x/interchainqueries/types"
	ictxkeeper "github.com/neutron-org/neutron/v4/x/interchaintxs/keeper"

	tokenfactorytypes "github.com/neutron-org/neutron/v4/x/tokenfactory/types"
)

const FeeCollectorAddress = "neutron1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7"

type CustomMessengerTestSuite struct {
	testutil.IBCConnectionTestSuite
	neutron         *app.App
	ctx             sdk.Context
	messenger       *wasmbinding.CustomMessenger
	contractOwner   sdk.AccAddress
	contractAddress sdk.AccAddress
	contractKeeper  wasmtypes.ContractOpsKeeper
}

func (suite *CustomMessengerTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
	suite.neutron = suite.GetNeutronZoneApp(suite.ChainA)
	suite.ctx = suite.ChainA.GetContext()
	suite.messenger = &wasmbinding.CustomMessenger{}
	suite.messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(suite.neutron.InterchainTxsKeeper)
	suite.messenger.Keeper = suite.neutron.InterchainTxsKeeper
	suite.messenger.Icqmsgserver = icqkeeper.NewMsgServerImpl(suite.neutron.InterchainQueriesKeeper)
	suite.messenger.Adminserver = adminkeeper.NewMsgServerImpl(suite.neutron.AdminmoduleKeeper)
	suite.messenger.Bank = &suite.neutron.BankKeeper
	suite.messenger.TokenFactory = suite.neutron.TokenFactoryKeeper
	suite.messenger.CronKeeper = &suite.neutron.CronKeeper
	suite.messenger.AdminKeeper = &suite.neutron.AdminmoduleKeeper
	suite.messenger.ContractmanagerKeeper = &suite.neutron.ContractManagerKeeper
	suite.contractOwner = keeper.RandomAccountAddress(suite.T())

	suite.contractKeeper = keeper.NewDefaultPermissionKeeper(&suite.neutron.WasmKeeper)

	err := suite.messenger.TokenFactory.SetParams(suite.ctx, tokenfactorytypes.NewParams(
		sdk.NewCoins(sdk.NewInt64Coin(params.DefaultDenom, 10_000_000)),
		0,
		FeeCollectorAddress,
	))
	suite.Require().NoError(err)

	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEmpty(suite.contractAddress)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainAccount() {
	err := suite.neutron.FeeBurnerKeeper.SetParams(suite.ctx, feeburnertypes.Params{
		NeutronDenom:    "untrn",
		TreasuryAddress: "neutron13jrwrtsyjjuynlug65r76r2zvfw5xjcq6532h2",
	})
	suite.Require().NoError(err)

	// Craft RegisterInterchainAccount message
	msg := bindings.NeutronMsg{
		RegisterInterchainAccount: &bindings.RegisterInterchainAccount{
			ConnectionId:        suite.Path.EndpointA.ConnectionID,
			InterchainAccountId: testutil.TestInterchainID,
			RegisterFee:         sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1_000_000))),
		},
	}

	bankKeeper := suite.neutron.BankKeeper
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	err = bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1_000_000))))
	suite.NoError(err)

	// Dispatch RegisterInterchainAccount message
	_, err = suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainAccountLongID() {
	// Craft RegisterInterchainAccount message
	msg, err := json.Marshal(bindings.NeutronMsg{
		RegisterInterchainAccount: &bindings.RegisterInterchainAccount{
			ConnectionId: suite.Path.EndpointA.ConnectionID,
			// the limit is 47, this line is 50 characters long
			InterchainAccountId: "01234567890123456789012345678901234567890123456789",
		},
	})
	suite.NoError(err)

	// Dispatch RegisterInterchainAccount message via DispatchHandler cause we want to catch an error from SDK directly, not from a contract
	_, _, _, err = suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{ //nolint:dogsled
		Custom: msg,
	})
	suite.Error(err)
	suite.ErrorIs(err, ictxtypes.ErrLongInterchainAccountID)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainQuery() {
	err := testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	// Top up contract balance
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(10_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	err = bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	// Craft RegisterInterchainQuery message
	msg := bindings.NeutronMsg{
		RegisterInterchainQuery: &bindings.RegisterInterchainQuery{
			QueryType: string(icqtypes.InterchainQueryTypeKV),
			Keys: []*icqtypes.KVKey{
				{Path: ibchost.StoreKey, Key: host.FullClientStateKey(suite.Path.EndpointB.ClientID)},
			},
			TransactionsFilter: "{}",
			ConnectionId:       suite.Path.EndpointA.ConnectionID,
			UpdatePeriod:       20,
		},
	}

	// Dispatch RegisterInterchainQuery message
	_, err = suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)
	suite.Equal(uint64(1), suite.neutron.InterchainQueriesKeeper.GetLastRegisteredQueryKey(suite.ctx))
}

func (suite *CustomMessengerTestSuite) TestCreateDenomMsg() {
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(10_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	fullMsg := bindings.NeutronMsg{
		CreateDenom: &bindings.CreateDenom{
			Subdenom: "SUN",
		},
	}

	data, err := suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)
	suite.Empty(data)
}

func (suite *CustomMessengerTestSuite) TestSetDenomMetadataMsg() {
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(10_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	fullMsg := bindings.NeutronMsg{
		CreateDenom: &bindings.CreateDenom{
			Subdenom: "SUN",
		},
	}
	data, err := suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.Empty(data)
	suite.NoError(err)

	sunDenom := fmt.Sprintf("factory/%s/%s", suite.contractAddress.String(), fullMsg.CreateDenom.Subdenom)
	metadata := banktypes.Metadata{
		Description: "very nice description",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    sunDenom,
				Exponent: 0,
				Aliases:  []string{"sun"},
			},
		},
		Base:    sunDenom,
		Display: sunDenom,
		Name:    "noname",
		Symbol:  sunDenom,
		URI:     "yuri",
		URIHash: "sdjalkfjsdklfj",
	}

	// this is the metadata variable but just in JSON representation, cause we can't marshal into json because of omitempty tags
	metaMsgBz := `
{
  "set_denom_metadata": {
    "description": "very nice description",
    "denom_units": [
      {
		"denom": "factory/neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq/SUN",
        "exponent": 0,
        "aliases": [
          "sun"
        ]
      }
    ],
    "base": "factory/neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq/SUN",
    "display": "factory/neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq/SUN",
    "name": "noname",
    "symbol": "factory/neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq/SUN",
    "uri": "yuri",
    "uri_hash": "sdjalkfjsdklfj"
  }
}
	`

	data, err = suite.executeCustomMsg(suite.contractAddress, json.RawMessage(metaMsgBz))
	suite.Empty(data)
	suite.NoError(err)

	metaFromBank, ok := bankKeeper.GetDenomMetaData(suite.ctx, sunDenom)
	suite.Require().True(ok)
	suite.Equal(metadata, metaFromBank)
}

func (suite *CustomMessengerTestSuite) TestMintMsg() {
	var (
		neutron = suite.GetNeutronZoneApp(suite.ChainA)
		ctx     = suite.ChainA.GetContext()
		lucky   = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
	)

	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(20_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	// lucky was broke
	balances := neutron.BankKeeper.GetAllBalances(suite.ctx, lucky)
	require.Empty(suite.T(), balances)

	// Create denom for minting
	fullMsg := bindings.NeutronMsg{
		CreateDenom: &bindings.CreateDenom{
			Subdenom: "SUN",
		},
	}

	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	sunDenom := fmt.Sprintf("factory/%s/%s", suite.contractAddress.String(), fullMsg.CreateDenom.Subdenom)

	amount, ok := math.NewIntFromString("808010808")
	require.True(suite.T(), ok)

	fullMsg = bindings.NeutronMsg{
		MintTokens: &bindings.MintTokens{
			Denom:         sunDenom,
			Amount:        amount,
			MintToAddress: lucky.String(),
		},
	}

	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	balances = neutron.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(suite.T(), balances, 1)
	coin := balances[0]
	require.Equal(suite.T(), amount, coin.Amount)
	require.Contains(suite.T(), coin.Denom, "factory/")
	require.Equal(suite.T(), sunDenom, coin.Denom)

	// mint the same denom again
	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	balances = neutron.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(suite.T(), balances, 1)
	coin = balances[0]
	require.Equal(suite.T(), amount.MulRaw(2), coin.Amount)
	require.Contains(suite.T(), coin.Denom, "factory/")
	require.Equal(suite.T(), sunDenom, coin.Denom)

	// now mint another amount / denom
	// create it first
	fullMsg = bindings.NeutronMsg{
		CreateDenom: &bindings.CreateDenom{
			Subdenom: "MOON",
		},
	}
	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	moonDenom := fmt.Sprintf("factory/%s/%s", suite.contractAddress.String(), fullMsg.CreateDenom.Subdenom)

	amount = amount.SubRaw(1)
	fullMsg = bindings.NeutronMsg{
		MintTokens: &bindings.MintTokens{
			Denom:         moonDenom,
			Amount:        amount,
			MintToAddress: lucky.String(),
		},
	}

	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	balances = neutron.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(suite.T(), balances, 2)
	coin = balances[0]
	require.Equal(suite.T(), amount, coin.Amount)
	require.Contains(suite.T(), coin.Denom, "factory/")
	require.Equal(suite.T(), moonDenom, coin.Denom)
}

func (suite *CustomMessengerTestSuite) TestForceTransferMsg() {
	var (
		neutron       = suite.GetNeutronZoneApp(suite.ChainA)
		ctx           = suite.ChainA.GetContext()
		lucky         = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
		forceReceiver = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
	)

	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(20_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	// lucky was broke
	balances := neutron.BankKeeper.GetAllBalances(suite.ctx, lucky)
	require.Empty(suite.T(), balances)

	// Create denom for minting
	fullMsg := bindings.NeutronMsg{
		CreateDenom: &bindings.CreateDenom{
			Subdenom: "SUN",
		},
	}

	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	sunDenom := fmt.Sprintf("factory/%s/%s", suite.contractAddress.String(), fullMsg.CreateDenom.Subdenom)

	amount, ok := math.NewIntFromString("808010808")
	require.True(suite.T(), ok)

	fullMsg = bindings.NeutronMsg{
		MintTokens: &bindings.MintTokens{
			Denom:         sunDenom,
			Amount:        amount,
			MintToAddress: lucky.String(),
		},
	}

	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	balances = neutron.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(suite.T(), balances, 1)
	coin := balances[0]
	require.Equal(suite.T(), amount, coin.Amount)
	require.Contains(suite.T(), coin.Denom, "factory/")
	require.Equal(suite.T(), sunDenom, coin.Denom)

	// now perform a force transfer to transfer tokens from a lucky address to a forceReceiver
	fullMsg = bindings.NeutronMsg{
		ForceTransfer: &bindings.ForceTransfer{
			Denom:               sunDenom,
			Amount:              amount,
			TransferFromAddress: lucky.String(),
			TransferToAddress:   forceReceiver.String(),
		},
	}

	_, err = suite.executeNeutronMsg(suite.contractAddress, fullMsg)
	suite.NoError(err)

	balancesLucky := neutron.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(suite.T(), balancesLucky, 0)
	balancesReceiver := neutron.BankKeeper.GetAllBalances(ctx, forceReceiver)
	require.Len(suite.T(), balancesReceiver, 1)

	coinReceiver := balancesReceiver[0]
	require.Equal(suite.T(), amount, coinReceiver.Amount)
	require.Contains(suite.T(), coinReceiver.Denom, "factory/")
	require.Equal(suite.T(), sunDenom, coinReceiver.Denom)
}

func (suite *CustomMessengerTestSuite) TestUpdateInterchainQuery() {
	// reuse register interchain query test to get query registered
	suite.TestRegisterInterchainQuery()

	// Craft UpdateInterchainQuery message
	msg := bindings.NeutronMsg{
		UpdateInterchainQuery: &bindings.UpdateInterchainQuery{
			QueryId:         1,
			NewKeys:         nil,
			NewUpdatePeriod: 111,
		},
	}

	// Dispatch UpdateInterchainQuery message
	_, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)
}

func (suite *CustomMessengerTestSuite) TestUpdateInterchainQueryFailed() {
	// Craft UpdateInterchainQuery message
	msg := bindings.NeutronMsg{
		UpdateInterchainQuery: &bindings.UpdateInterchainQuery{
			QueryId:         1,
			NewKeys:         nil,
			NewUpdatePeriod: 1,
		},
	}

	// Dispatch UpdateInterchainQuery message
	data, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	expectedErrMsg := "failed to update interchain query: failed to update interchain query: failed to get query by query id: there is no query with id: 1"
	suite.Require().ErrorContains(err, expectedErrMsg)
	suite.Empty(data)
}

func (suite *CustomMessengerTestSuite) TestRemoveInterchainQuery() {
	// Reuse register interchain query test to get query registered
	suite.TestRegisterInterchainQuery()

	// Craft RemoveInterchainQuery message
	msg := bindings.NeutronMsg{
		RemoveInterchainQuery: &bindings.RemoveInterchainQuery{
			QueryId: 1,
		},
	}

	// Dispatch RemoveInterchainQuery message
	_, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)
}

func (suite *CustomMessengerTestSuite) TestRemoveInterchainQueryFailed() {
	// Craft RemoveInterchainQuery message
	msg := bindings.NeutronMsg{
		RemoveInterchainQuery: &bindings.RemoveInterchainQuery{
			QueryId: 1,
		},
	}

	// Dispatch RemoveInterchainQuery message
	data, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	expectedErrMsg := "failed to remove interchain query: failed to remove interchain query: failed to get query by query id: there is no query with id: 1"
	suite.Require().ErrorContains(err, expectedErrMsg)
	suite.Empty(data)
}

func (suite *CustomMessengerTestSuite) TestSubmitTx() {
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(10_000_000))))
	bankKeeper := suite.neutron.BankKeeper
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	err = testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	data, err := suite.executeNeutronMsg(
		suite.contractAddress,
		suite.craftMarshaledMsgSubmitTxWithNumMsgs(1),
	)
	suite.NoError(err)

	var response ictxtypes.MsgSubmitTxResponse
	err = json.Unmarshal(data, &response)
	suite.NoError(err)
	suite.Equal(uint64(1), response.SequenceId)
	suite.Equal("channel-2", response.Channel)
}

func (suite *CustomMessengerTestSuite) TestSubmitTxTooMuchTxs() {
	err := testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	_, err = suite.executeNeutronMsg(
		suite.contractAddress,
		suite.craftMarshaledMsgSubmitTxWithNumMsgs(20),
	)
	suite.ErrorContains(err, "MsgSubmitTx contains more messages than allowed")
}

func (suite *CustomMessengerTestSuite) TestSoftwareUpgradeProposal() {
	// Set admin so that we can execute this proposal without permission error
	suite.neutron.AdminmoduleKeeper.SetAdmin(suite.ctx, suite.contractAddress.String())

	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	anotherContract := suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEqual(anotherContract, suite.contractAddress)

	executeMsg := fmt.Sprintf(`
{
  "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
  "authority": "%s",
  "plan": {
    "name": "TestPlane",
    "height": "150",
    "info": "TestInfo"
  }
}
`, suite.neutron.BankKeeper.GetAuthority())
	// Craft SubmitAdminProposal message
	msg := bindings.NeutronMsg{
		SubmitAdminProposal: &bindings.SubmitAdminProposal{
			AdminProposal: bindings.AdminProposal{
				ProposalExecuteMessage: &bindings.ProposalExecuteMessage{
					Message: executeMsg,
				},
			},
		},
	}

	// Dispatch SubmitAdminProposal message
	data, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)

	expected, err := json.Marshal(&admintypes.MsgSubmitProposalResponse{
		ProposalId: 1,
	})
	suite.NoError(err)
	suite.Equal(expected, data)

	// Test with other proposer that is not admin should return failure
	_, err = suite.executeNeutronMsg(anotherContract, msg)
	suite.Error(err)

	// Check CancelSubmitAdminProposal

	executeMsg = fmt.Sprintf(`
				{
                "@type": "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
                "authority": "%s"
}
`, suite.neutron.BankKeeper.GetAuthority())
	// Craft CancelSubmitAdminProposal message
	msg = bindings.NeutronMsg{
		SubmitAdminProposal: &bindings.SubmitAdminProposal{
			AdminProposal: bindings.AdminProposal{ProposalExecuteMessage: &bindings.ProposalExecuteMessage{Message: executeMsg}},
		},
	}

	// Dispatch SubmitAdminProposal message
	data, err = suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)

	expected, err = json.Marshal(&admintypes.MsgSubmitProposalResponse{
		ProposalId: 2,
	})
	suite.NoError(err)
	suite.Equal(expected, data)
}

func (suite *CustomMessengerTestSuite) TestTooMuchProposals() {
	// Store code and instantiate reflect contract
	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEmpty(suite.contractAddress)

	err := testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	executeMsg := fmt.Sprintf(`
				{
                "@type": "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
                "authority": "%s"
}
`, suite.neutron.BankKeeper.GetAuthority())

	// Craft  message with 2 proposals
	msg, err := json.Marshal(bindings.NeutronMsg{
		SubmitAdminProposal: &bindings.SubmitAdminProposal{
			AdminProposal: bindings.AdminProposal{
				ParamChangeProposal: &bindings.ParamChangeProposal{
					Title:        "aaa",
					Description:  "ddafds",
					ParamChanges: nil,
				},
				ProposalExecuteMessage: &bindings.ProposalExecuteMessage{Message: executeMsg},
			},
		},
	})
	suite.NoError(err)

	cosmosMsg := types.CosmosMsg{Custom: msg}

	// Dispatch SubmitAdminProposal message
	_, _, _, err = suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, cosmosMsg) //nolint:dogsled

	suite.ErrorContains(err, "more than one admin proposal type is present in message")
}

func (suite *CustomMessengerTestSuite) TestNoProposals() {
	err := testutil.SetupICAPath(suite.Path, suite.contractAddress.String())
	suite.Require().NoError(err)

	// Craft  message with 0 proposals
	msg, err := json.Marshal(bindings.NeutronMsg{
		SubmitAdminProposal: &bindings.SubmitAdminProposal{
			AdminProposal: bindings.AdminProposal{},
		},
	})
	suite.NoError(err)

	cosmosMsg := types.CosmosMsg{Custom: msg}

	// Dispatch SubmitAdminProposal message
	_, _, _, err = suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, cosmosMsg) //nolint:dogsled

	suite.ErrorContains(err, "no admin proposal type is present in message")
}

func (suite *CustomMessengerTestSuite) TestAddRemoveSchedule() {
	// Set admin so that we can execute this proposal without permission error
	suite.neutron.AdminmoduleKeeper.SetAdmin(suite.ctx, suite.contractAddress.String())

	// Craft AddSchedule message
	msg := bindings.NeutronMsg{
		AddSchedule: &bindings.AddSchedule{
			Name:   "schedule1",
			Period: 5,
			Msgs: []bindings.MsgExecuteContract{
				{
					Contract: suite.contractAddress.String(),
					Msg:      "{\"send\": { \"to\": \"asdf\", \"amount\": 1000 }}",
				},
			},
		},
	}

	// Dispatch AddSchedule message
	_, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)

	// Craft RemoveSchedule message
	msg = bindings.NeutronMsg{
		RemoveSchedule: &bindings.RemoveSchedule{
			Name: "schedule1",
		},
	}

	// Dispatch AddSchedule message
	_, err = suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)
}

func (suite *CustomMessengerTestSuite) TestResubmitFailureAck() {
	// Add failure
	packet := ibcchanneltypes.Packet{}
	ack := ibcchanneltypes.Acknowledgement{
		Response: &ibcchanneltypes.Acknowledgement_Result{Result: []byte("Result")},
	}
	payload, err := keeper2.PrepareSudoCallbackMessage(packet, &ack)
	require.NoError(suite.T(), err)
	failureID := suite.messenger.ContractmanagerKeeper.GetNextFailureIDKey(suite.ctx, suite.contractAddress.String())
	suite.messenger.ContractmanagerKeeper.AddContractFailure(suite.ctx, suite.contractAddress.String(), payload, "test error")

	// Craft message
	msg := bindings.NeutronMsg{
		ResubmitFailure: &bindings.ResubmitFailure{
			FailureId: failureID,
		},
	}

	// Dispatch
	data, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)

	expected, err := json.Marshal(&bindings.ResubmitFailureResponse{})
	suite.NoError(err)
	suite.Equal(expected, data)
}

func (suite *CustomMessengerTestSuite) TestResubmitFailureTimeout() {
	// Add failure
	packet := ibcchanneltypes.Packet{}
	payload, err := keeper2.PrepareSudoCallbackMessage(packet, nil)
	require.NoError(suite.T(), err)
	failureID := suite.messenger.ContractmanagerKeeper.GetNextFailureIDKey(suite.ctx, suite.contractAddress.String())
	suite.messenger.ContractmanagerKeeper.AddContractFailure(suite.ctx, suite.contractAddress.String(), payload, "test error")

	// Craft message
	msg := bindings.NeutronMsg{
		ResubmitFailure: &bindings.ResubmitFailure{
			FailureId: failureID,
		},
	}

	// Dispatch
	data, err := suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.NoError(err)

	expected, err := json.Marshal(&bindings.ResubmitFailureResponse{FailureId: failureID})
	suite.NoError(err)
	suite.Equal(expected, data)
}

func (suite *CustomMessengerTestSuite) TestResubmitFailureFromDifferentContract() {
	// Add failure
	packet := ibcchanneltypes.Packet{}
	ack := ibcchanneltypes.Acknowledgement{
		Response: &ibcchanneltypes.Acknowledgement_Error{Error: "ErrorSudoPayload"},
	}
	failureID := suite.messenger.ContractmanagerKeeper.GetNextFailureIDKey(suite.ctx, testutil.TestOwnerAddress)
	payload, err := keeper2.PrepareSudoCallbackMessage(packet, &ack)
	require.NoError(suite.T(), err)
	suite.messenger.ContractmanagerKeeper.AddContractFailure(suite.ctx, testutil.TestOwnerAddress, payload, "test error")

	// Craft message
	msg := bindings.NeutronMsg{
		ResubmitFailure: &bindings.ResubmitFailure{
			FailureId: failureID,
		},
	}

	// Dispatch
	_, err = suite.executeNeutronMsg(suite.contractAddress, msg)
	suite.ErrorContains(err, "no failure found to resubmit: not found")
}

func (suite *CustomMessengerTestSuite) executeCustomMsg(contractAddress sdk.AccAddress, fullMsg json.RawMessage) (data []byte, err error) {
	customMsg := types.CosmosMsg{
		Custom: fullMsg,
	}

	type ExecuteMsg struct {
		ReflectMsg struct {
			Msgs []types.CosmosMsg `json:"msgs"`
		} `json:"reflect_msg"`
	}

	execMsg := ExecuteMsg{ReflectMsg: struct {
		Msgs []types.CosmosMsg `json:"msgs"`
	}(struct{ Msgs []types.CosmosMsg }{Msgs: []types.CosmosMsg{customMsg}})}

	msg, err := json.Marshal(execMsg)
	suite.NoError(err)

	data, err = suite.contractKeeper.Execute(suite.ctx, contractAddress, suite.contractOwner, msg, nil)

	return
}

func (suite *CustomMessengerTestSuite) executeNeutronMsg(contractAddress sdk.AccAddress, fullMsg bindings.NeutronMsg) (data []byte, err error) {
	fullMsgBz, err := json.Marshal(fullMsg)
	suite.NoError(err)

	return suite.executeCustomMsg(contractAddress, fullMsgBz)
}

func (suite *CustomMessengerTestSuite) craftMarshaledMsgSubmitTxWithNumMsgs(numMsgs int) bindings.NeutronMsg {
	msg := bindings.ProtobufAny{
		TypeURL: "/cosmos.staking.v1beta1.MsgDelegate",
		Value:   []byte{26, 10, 10, 5, 115, 116, 97, 107, 101, 18, 1, 48},
	}
	msgs := make([]bindings.ProtobufAny, 0, numMsgs)
	for i := 0; i < numMsgs; i++ {
		msgs = append(msgs, msg)
	}
	result := bindings.NeutronMsg{
		SubmitTx: &bindings.SubmitTx{
			ConnectionId:        suite.Path.EndpointA.ConnectionID,
			InterchainAccountId: testutil.TestInterchainID,
			Msgs:                msgs,
			Memo:                "Jimmy",
			Timeout:             2000,
			Fee: feetypes.Fee{
				RecvFee:    sdk.NewCoins(),
				AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1000))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(1000))),
			},
		},
	}
	return result
}

func TestMessengerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomMessengerTestSuite))
}
