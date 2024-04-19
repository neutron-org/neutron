package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/baseapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/app/params"
	"github.com/neutron-org/neutron/v3/testutil"
	"github.com/neutron-org/neutron/v3/x/tokenfactory/keeper"
	"github.com/neutron-org/neutron/v3/x/tokenfactory/types"
)

const (
	FeeCollectorAddress = "neutron1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7"
	TopUpCoinsAmount    = 1_000_000
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite

	TestAccs    []sdktypes.AccAddress
	QueryHelper *baseapp.QueryServiceTestHelper

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
	))
	suite.Require().NoError(err)

	suite.msgServer = keeper.NewMsgServerImpl(*tokenFactoryKeeper)
}

func (suite *KeeperTestSuite) SetupTokenFactory() {
	suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.CreateModuleAccount(suite.ChainA.GetContext())
}

func (suite *KeeperTestSuite) CreateDefaultDenom(ctx sdktypes.Context) {
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(ctx, senderAddress, suite.TestAccs[0])

	res, _ := suite.msgServer.CreateDenom(sdktypes.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.defaultDenom = res.GetNewTokenDenom()
}

func (suite *KeeperTestSuite) TopUpWallet(ctx sdktypes.Context, sender, contractAddress sdktypes.AccAddress) {
	coinsAmnt := sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(TopUpCoinsAmount)))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) WalletBalance(ctx sdktypes.Context, address string) math.Int {
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	balance, err := bankKeeper.Balance(
		sdktypes.WrapSDKContext(ctx),
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

func (s *KeeperTestSuite) TestForceTransferMsg() {
	s.Setup()

	// Create a denom
	s.CreateDefaultDenom(s.ChainA.GetContext())

	s.Run(fmt.Sprintf("test force transfer"), func() {
		mintAmt := sdktypes.NewInt64Coin(s.defaultDenom, 10)

		_, err := s.msgServer.Mint(sdktypes.WrapSDKContext(s.ChainA.GetContext()), types.NewMsgMint(s.TestAccs[0].String(), mintAmt))

		govModAcc := s.GetNeutronZoneApp(s.ChainA).AccountKeeper.GetModuleAccount(s.ChainA.GetContext(), authtypes.FeeCollectorName)

		err = s.GetNeutronZoneApp(s.ChainA).BankKeeper.SendCoins(s.ChainA.GetContext(), s.TestAccs[0], govModAcc.GetAddress(), sdktypes.NewCoins(mintAmt))
		s.Require().NoError(err)

		_, err = s.msgServer.ForceTransfer(s.ChainA.GetContext(), types.NewMsgForceTransfer(s.TestAccs[0].String(), mintAmt, govModAcc.GetAddress().String(), s.TestAccs[1].String()))
		s.Require().ErrorContains(err, "send from module acc not available")
	})
}
