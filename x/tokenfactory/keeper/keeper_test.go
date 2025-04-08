package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/baseapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/app/params"
	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/keeper"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

const (
	FeeCollectorAddress = "neutron1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7"
	TopUpCoinsAmount    = 1_000_000
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite

	TestAccs    []sdktypes.AccAddress
	QueryHelper *baseapp.QueryServiceTestHelper

	contractKeeper wasmtypes.ContractOpsKeeper

	bankMsgServer banktypes.MsgServer

	queryClient types.QueryClient
	msgServer   types.MsgServer
	// defaultDenom is on the suite, as it depends on the creator test address.
	defaultDenom string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) Setup() {
	suite.SetupTest()

	suite.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: suite.GetNeutronZoneApp(suite.ChainA).GRPCQueryRouter(),
		Ctx:             suite.ChainA.GetContext(),
	}
	suite.TestAccs = CreateRandomAccounts(3)

	suite.SetupTokenFactory()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	tokenFactoryKeeper := suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper
	err := tokenFactoryKeeper.SetParams(suite.ChainA.GetContext(), types.NewParams(
		sdktypes.NewCoins(sdktypes.NewInt64Coin(params.DefaultDenom, TopUpCoinsAmount)),
		0,
		FeeCollectorAddress,
		types.DefaultWhitelistedHooks,
	))
	suite.Require().NoError(err)

	suite.bankMsgServer = bankkeeper.NewMsgServerImpl(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper)
	suite.msgServer = keeper.NewMsgServerImpl(*tokenFactoryKeeper)

	suite.contractKeeper = wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(suite.ChainA).WasmKeeper)
}

func (suite *KeeperTestSuite) SetupTokenFactory() {
	suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.CreateModuleAccount(suite.ChainA.GetContext())
}

func (suite *KeeperTestSuite) CreateDefaultDenom(ctx sdktypes.Context) {
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(ctx, senderAddress, suite.TestAccs[0])

	res, _ := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.defaultDenom = res.GetNewTokenDenom()
}

func (suite *KeeperTestSuite) TopUpWallet(ctx sdktypes.Context, sender, contractAddress sdktypes.AccAddress) {
	coinsAmnt := sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, math.NewInt(TopUpCoinsAmount)))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) WalletBalance(ctx sdktypes.Context, address string) math.Int {
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	balance, err := bankKeeper.Balance(
		ctx,
		&banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   params.DefaultDenom,
		},
	)
	suite.Require().NoError(err)

	return balance.Balance.Amount
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdktypes.AccAddress {
	testAddrs := make([]sdktypes.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdktypes.AccAddress(pk.Address())
	}

	return testAddrs
}

func (suite *KeeperTestSuite) TestForceTransferMsg() {
	suite.Setup()

	// Create a denom
	suite.CreateDefaultDenom(suite.ChainA.GetContext())

	suite.Run("test force transfer", func() {
		mintAmt := sdktypes.NewInt64Coin(suite.defaultDenom, 10)

		_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), mintAmt))
		suite.Require().NoError(err)

		govModAcc := suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAccount(suite.ChainA.GetContext(), authtypes.FeeCollectorName)

		err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoins(suite.ChainA.GetContext(), suite.TestAccs[0], govModAcc.GetAddress(), sdktypes.NewCoins(mintAmt))
		suite.Require().NoError(err)

		_, err = suite.msgServer.ForceTransfer(suite.ChainA.GetContext(), types.NewMsgForceTransfer(suite.TestAccs[0].String(), mintAmt, govModAcc.GetAddress().String(), suite.TestAccs[1].String()))
		suite.Require().ErrorContains(err, "force transfer from module accounts is forbidden")

		_, err = suite.msgServer.ForceTransfer(suite.ChainA.GetContext(), types.NewMsgForceTransfer(suite.TestAccs[0].String(), mintAmt, suite.TestAccs[1].String(), govModAcc.GetAddress().String()))
		suite.Require().ErrorContains(err, "force transfer to module accounts is forbidden")
	})
}

func (suite *KeeperTestSuite) TestMintToMsg() {
	suite.Setup()

	// Create a denom
	suite.CreateDefaultDenom(suite.ChainA.GetContext())

	suite.Run("test mint to", func() {
		mintAmt := sdktypes.NewInt64Coin(suite.defaultDenom, 10)

		govModAcc := suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAccount(suite.ChainA.GetContext(), authtypes.FeeCollectorName)

		_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMintTo(suite.TestAccs[0].String(), mintAmt, govModAcc.GetAddress().String()))
		suite.Require().ErrorContains(err, "minting to module accounts is forbidden")
	})
}

func (suite *KeeperTestSuite) TestBurnFromMsg() {
	suite.Setup()

	// Create a denom
	suite.CreateDefaultDenom(suite.ChainA.GetContext())

	suite.Run("test burn from", func() {
		mintAmt := sdktypes.NewInt64Coin(suite.defaultDenom, 10)

		_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), mintAmt))
		suite.Require().NoError(err)

		govModAcc := suite.GetNeutronZoneApp(suite.ChainA).AccountKeeper.GetModuleAccount(suite.ChainA.GetContext(), authtypes.FeeCollectorName)

		err = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.SendCoins(suite.ChainA.GetContext(), suite.TestAccs[0], govModAcc.GetAddress(), sdktypes.NewCoins(mintAmt))
		suite.Require().NoError(err)

		_, err = suite.msgServer.Burn(suite.ChainA.GetContext(), types.NewMsgBurnFrom(suite.TestAccs[0].String(), mintAmt, govModAcc.GetAddress().String()))
		suite.Require().ErrorContains(err, "burning from module accounts is forbidden")
	})
}
