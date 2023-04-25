package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/app/params"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/x/tokenfactory/keeper"
	"github.com/neutron-org/neutron/x/tokenfactory/types"
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

	tokeFactoryKeeper := suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper
	tokeFactoryKeeper.SetParams(suite.ChainA.GetContext(), types.NewParams(
		sdktypes.NewCoins(sdktypes.NewInt64Coin(types.DefaultNeutronDenom, TopUpCoinsAmount)),
		FeeCollectorAddress,
	))

	suite.msgServer = keeper.NewMsgServerImpl(*tokeFactoryKeeper)
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

func (suite *KeeperTestSuite) TopUpWallet(ctx sdktypes.Context, sender sdktypes.AccAddress, contractAddress sdktypes.AccAddress) {
	coinsAmnt := sdktypes.NewCoins(sdktypes.NewCoin(params.DefaultDenom, sdktypes.NewInt(TopUpCoinsAmount)))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) WalletBalance(ctx sdktypes.Context, address string) sdktypes.Int {
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
