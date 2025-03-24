package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/app/params"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestAdminMsgs() {
	addr0bal := int64(0)
	addr1bal := int64(0)

	suite.Setup()
	suite.CreateDefaultDenom(suite.ChainA.GetContext())
	// Make sure that the admin is set correctly
	denom := strings.Split(suite.defaultDenom, "/")
	queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.ChainA.GetContext().Context(), &types.QueryDenomAuthorityMetadataRequest{
		Creator:  denom[1],
		Subdenom: denom[2],
	})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

	// Test minting to admins own account
	_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(suite.defaultDenom, 10)))
	addr0bal += 10
	suite.Require().NoError(err)
	suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64() == addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))

	// Test burning from own account
	_, err = suite.msgServer.Burn(suite.ChainA.GetContext(), types.NewMsgBurn(suite.TestAccs[0].String(), sdk.NewInt64Coin(suite.defaultDenom, 5)))
	addr0bal -= 5
	suite.Require().NoError(err)
	suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64() == addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))
	suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[1], suite.defaultDenom).Amount.Int64() == addr1bal)

	// Test Change Admin
	_, err = suite.msgServer.ChangeAdmin(suite.ChainA.GetContext(), types.NewMsgChangeAdmin(suite.TestAccs[0].String(), suite.defaultDenom, suite.TestAccs[1].String()))
	suite.Require().NoError(err)
	denom = strings.Split(suite.defaultDenom, "/")
	queryRes, err = suite.queryClient.DenomAuthorityMetadata(suite.ChainA.GetContext().Context(), &types.QueryDenomAuthorityMetadataRequest{
		Creator:  denom[1],
		Subdenom: denom[2],
	})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[1].String(), queryRes.AuthorityMetadata.Admin)

	// Make sure old admin can no longer do actions
	_, err = suite.msgServer.Burn(suite.ChainA.GetContext(), types.NewMsgBurn(suite.TestAccs[0].String(), sdk.NewInt64Coin(suite.defaultDenom, 5)))
	suite.Require().Error(err)

	// Make sure the new admin works
	_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[1].String(), sdk.NewInt64Coin(suite.defaultDenom, 5)))
	addr1bal += 5
	suite.Require().NoError(err)
	suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[1], suite.defaultDenom).Amount.Int64() == addr1bal)

	// Try setting admin to empty
	_, err = suite.msgServer.ChangeAdmin(suite.ChainA.GetContext(), types.NewMsgChangeAdmin(suite.TestAccs[1].String(), suite.defaultDenom, ""))
	suite.Require().Error(err)
}

// TestMintDenom ensures the following properties of the MintMessage:
// * Noone can mint tokens for a denom that doesn't exist
// * Only the admin of a denom can mint tokens for it
// * The admin of a denom can mint tokens for it
func (suite *KeeperTestSuite) TestMintDenom() {
	addr0bal := int64(0)
	suite.Setup()

	// Create a denom
	suite.CreateDefaultDenom(suite.ChainA.GetContext())

	for _, tc := range []struct {
		desc      string
		amount    int64
		mintDenom string
		admin     string
		valid     bool
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			mintDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     suite.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:      "mint is not by the admin",
			amount:    10,
			mintDenom: suite.defaultDenom,
			admin:     suite.TestAccs[1].String(),
			valid:     false,
		},
		{
			desc:      "success case",
			amount:    10,
			mintDenom: suite.defaultDenom,
			admin:     suite.TestAccs[0].String(),
			valid:     true,
		},
		{
			desc:      "error: try minting non-tokenfactory denom",
			amount:    10,
			mintDenom: params.DefaultDenom,
			admin:     suite.TestAccs[0].String(),
			valid:     false,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// Test minting to admins own account
			_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(tc.admin, sdk.NewInt64Coin(tc.mintDenom, tc.amount)))
			if tc.valid {
				addr0bal += 10
				suite.Require().NoError(err)
				suite.Require().Equal(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64(), addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBurnDenom() {
	addr0bal := int64(0)
	suite.Setup()

	// Create a denom.
	suite.CreateDefaultDenom(suite.ChainA.GetContext())

	// mint 10 default token for testAcc[0]
	_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(suite.defaultDenom, 10)))
	suite.NoError(err)
	addr0bal += 10

	for _, tc := range []struct {
		desc      string
		amount    int64
		burnDenom string
		admin     string
		valid     bool
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			burnDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     suite.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:      "burn is not by the admin",
			amount:    10,
			burnDenom: suite.defaultDenom,
			admin:     suite.TestAccs[1].String(),
			valid:     false,
		},
		{
			desc:      "burn amount is bigger than minted amount",
			amount:    1000,
			burnDenom: suite.defaultDenom,
			admin:     suite.TestAccs[1].String(),
			valid:     false,
		},
		{
			desc:      "success case",
			amount:    10,
			burnDenom: suite.defaultDenom,
			admin:     suite.TestAccs[0].String(),
			valid:     true,
		},
		{
			desc:      "fail case - burn non-tokenfactory denom",
			amount:    10,
			burnDenom: params.DefaultDenom,
			admin:     suite.TestAccs[0].String(),
			valid:     false,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// Test minting to admins own account
			_, err := suite.msgServer.Burn(suite.ChainA.GetContext(), types.NewMsgBurn(tc.admin, sdk.NewInt64Coin(tc.burnDenom, tc.amount)))
			if tc.valid {
				addr0bal -= 10
				suite.Require().NoError(err)
				suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64() == addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))
			} else {
				suite.Require().Error(err)
				suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64() == addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestForceTransferDenom() {
	addr0bal := int64(0)
	addr1bal := int64(0)
	suite.Setup()

	// Create a denom.
	suite.CreateDefaultDenom(suite.ChainA.GetContext())

	// mint 10 default token for testAcc[0]
	_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(suite.defaultDenom, 10)))
	suite.NoError(err)
	addr0bal += 10

	for _, tc := range []struct {
		desc          string
		amount        int64
		transferDenom string
		admin         string
		valid         bool
	}{
		{
			desc:          "denom does not exist",
			amount:        10,
			transferDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:         suite.TestAccs[0].String(),
			valid:         false,
		},
		{
			desc:          "force transfer is not by the admin",
			amount:        10,
			transferDenom: suite.defaultDenom,
			admin:         suite.TestAccs[1].String(),
			valid:         false,
		},
		{
			desc:          "force transfer amount is bigger than minted amount",
			amount:        1000,
			transferDenom: suite.defaultDenom,
			admin:         suite.TestAccs[0].String(),
			valid:         false,
		},
		{
			desc:          "success case",
			amount:        10,
			transferDenom: suite.defaultDenom,
			admin:         suite.TestAccs[0].String(),
			valid:         true,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// Test minting to admins own account
			_, err := suite.msgServer.ForceTransfer(suite.ChainA.GetContext(), types.NewMsgForceTransfer(tc.admin, sdk.NewInt64Coin(tc.transferDenom, tc.amount), suite.TestAccs[0].String(), suite.TestAccs[1].String()))
			if tc.valid {
				addr0bal -= 10
				addr1bal += 10

				suite.Require().NoError(err)
				suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64() == addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))
				suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[1], suite.defaultDenom).Amount.Int64() == addr1bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[1], suite.defaultDenom))
			} else {
				suite.Require().Error(err)
				suite.Require().True(suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom).Amount.Int64() == addr0bal, suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), suite.TestAccs[0], suite.defaultDenom))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChangeAdminDenom() {
	for _, tc := range []struct {
		desc                    string
		msgChangeAdmin          func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminPass bool
		expectedAdminIndex      int
		msgMint                 func(denom string) *types.MsgMint
		expectedMintPass        bool
	}{
		{
			desc: "can't set admin to '' ",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(suite.TestAccs[0].String(), denom, "")
			},
			expectedChangeAdminPass: false,
			expectedAdminIndex:      0,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: true,
		},
		{
			desc: "non-admins can't change the existing admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(suite.TestAccs[1].String(), denom, suite.TestAccs[2].String())
			},
			expectedChangeAdminPass: false,
			expectedAdminIndex:      0,
		},
		{
			desc: "success change admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(suite.TestAccs[0].String(), denom, suite.TestAccs[1].String())
			},
			expectedAdminIndex:      1,
			expectedChangeAdminPass: true,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(suite.TestAccs[1].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: true,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.Setup()

			// Create a denom and mint
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(suite.ChainA.GetContext(), senderAddress, suite.TestAccs[0])

			res, err := suite.msgServer.CreateDenom(suite.ChainA.GetContext(), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
			suite.Require().NoError(err)

			testDenom := res.GetNewTokenDenom()

			_, err = suite.msgServer.Mint(suite.ChainA.GetContext(), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(testDenom, 10)))
			suite.Require().NoError(err)

			_, err = suite.msgServer.ChangeAdmin(suite.ChainA.GetContext(), tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			denom := strings.Split(testDenom, "/")
			queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.ChainA.GetContext().Context(), &types.QueryDenomAuthorityMetadataRequest{
				Creator:  denom[1],
				Subdenom: denom[2],
			})
			suite.Require().NoError(err)

			// expectedAdminIndex with negative value is assumed as admin with value of ""
			const emptyStringAdminIndexFlag = -1
			if tc.expectedAdminIndex == emptyStringAdminIndexFlag {
				suite.Require().Equal("", queryRes.AuthorityMetadata.Admin)
			} else {
				suite.Require().Equal(suite.TestAccs[tc.expectedAdminIndex].String(), queryRes.AuthorityMetadata.Admin)
			}

			// we test mint to test if admin authority is performed properly after admin change.
			if tc.msgMint != nil {
				_, err := suite.msgServer.Mint(suite.ChainA.GetContext(), tc.msgMint(testDenom))
				if tc.expectedMintPass {
					suite.Require().NoError(err)
				} else {
					suite.Require().Error(err)
				}
			}
		})
	}
}
