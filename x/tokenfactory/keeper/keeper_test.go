package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/x/tokenfactory/keeper"
	"github.com/neutron-org/neutron/x/tokenfactory/types"
)

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite

	TestAccs    []sdk.AccAddress
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
	fmt.Printf("Setup: %s\n", suite.TestAccs)

	// Fund every TestAcc with 100 denom creation fees.
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, fundAccsAmount)
	}

	suite.SetupTokenFactory()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	tokeFactoryKeeper := suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper
	suite.msgServer = keeper.NewMsgServerImpl(*tokeFactoryKeeper)
}

func (suite *KeeperTestSuite) SetupTokenFactory() {
	suite.GetNeutronZoneApp(suite.ChainA).TokenFactoryKeeper.CreateModuleAccount(suite.ChainA.GetContext())
}

func (suite *KeeperTestSuite) CreateDefaultDenom() {
	fmt.Printf("CreateDefaultDenom: %s\n", suite.TestAccs)
	res, _ := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.ChainA.GetContext()), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.defaultDenom = res.GetNewTokenDenom()
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}
