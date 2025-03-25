package keeper_test

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestMsgCreateDenom() {
	suite.Setup()

	// Create denom without enough funds
	_, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.Require().ErrorIs(err, errors.ErrInsufficientFunds)

	// Creating a denom should work
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

	balance := suite.WalletBalance(suite.ChainA.GetContext(), suite.TestAccs[0].String())
	suite.Require().Equal(math.NewInt(TopUpCoinsAmount), balance)

	feeCollectorBalance := suite.WalletBalance(suite.ChainA.GetContext(), FeeCollectorAddress)
	suite.Require().Equal(math.NewInt(0), feeCollectorBalance)

	res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(res.GetNewTokenDenom())

	balance = suite.WalletBalance(suite.ChainA.GetContext(), suite.TestAccs[0].String())
	suite.Require().Equal(math.NewInt(0), balance)

	feeCollectorBalance = suite.WalletBalance(suite.ChainA.GetContext(), FeeCollectorAddress)
	suite.Require().Equal(math.NewInt(TopUpCoinsAmount), feeCollectorBalance)

	// Make sure that the admin is set correctly
	denom := strings.Split(res.GetNewTokenDenom(), "/")

	queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.ChainA.GetContext().Context(), &types.QueryDenomAuthorityMetadataRequest{
		Creator:  denom[1],
		Subdenom: denom[2],
	})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

	// Make sure that a second version of the same denom can't be recreated
	res, err = suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.Empty(res)
	suite.Require().Error(err)

	// Creating a second denom should work
	suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

	res, err = suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "litecoin"))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(res.GetNewTokenDenom())

	// Try querying all the denoms created by suite.TestAccs[0]
	queryRes2, err := suite.queryClient.DenomsFromCreator(suite.ChainA.GetContext().Context(), &types.QueryDenomsFromCreatorRequest{
		Creator: suite.TestAccs[0].String(),
	})
	suite.Require().NoError(err)
	suite.Require().Len(queryRes2.Denoms, 2)

	// Make sure that a second account can create a denom with the same subdenom
	suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[1])

	res, err = suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[1].String(), "bitcoin"))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(res.GetNewTokenDenom())

	// Make sure that an address with a "/" in it can't create denoms
	res, err = suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom("osmosis.eth/creator", "bitcoin"))
	suite.Empty(res)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestCreateDenom() {
	for _, tc := range []struct {
		desc     string
		setup    func()
		subdenom string
		valid    bool
	}{
		{
			desc:     "subdenom too long",
			subdenom: "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:    false,
		},
		{
			desc: "subdenom and creator pair already exists",
			setup: func() {
				senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
				suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])
				_, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
				suite.Require().NoError(err)
			},
			subdenom: "bitcoin",
			valid:    false,
		},
		{
			desc:     "success case",
			subdenom: "evmos",
			valid:    true,
		},
		{
			desc:     "subdenom having invalid characters",
			subdenom: "bit/***///&&&/coin",
			valid:    false,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.Setup()

			if tc.setup != nil {
				tc.setup()
			}
			// Create a denom
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

			res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), tc.subdenom))
			if tc.valid {
				suite.Require().NoError(err)

				denom := strings.Split(res.GetNewTokenDenom(), "/")

				// Make sure that the admin is set correctly
				queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.ChainA.GetContext().Context(), &types.QueryDenomAuthorityMetadataRequest{
					Creator:  denom[1],
					Subdenom: denom[2],
				})

				suite.Require().NoError(err)
				suite.Require().Equal(suite.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
