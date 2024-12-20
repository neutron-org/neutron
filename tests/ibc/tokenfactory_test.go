package ibc_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	tftypes "github.com/neutron-org/neutron/v5/x/tokenfactory/types"
	"github.com/stretchr/testify/suite"
)

type TokenfactoryTestSuite struct {
	IBCTestSuite
}

func TestTokenfactoryTestSuite(t *testing.T) {
	suite.Run(t, new(TokenfactoryTestSuite))
}

func (s *TokenfactoryTestSuite) TestForceTransferFromIBCEscrow() {
	// Create token factory denom
	createDenomMsg := tftypes.NewMsgCreateDenom(s.neutronAddr.String(), "testtest")
	_, err := s.neutronChain.SendMsgs(createDenomMsg)
	s.Assert().NoError(err)

	// Derive full token factory denom
	denom := fmt.Sprintf("factory/%s/%s", createDenomMsg.Sender, createDenomMsg.Subdenom)

	// Mint denom to sender
	amount := sdk.NewCoin(denom, math.NewInt(10000000))
	mintMsg := tftypes.NewMsgMint(createDenomMsg.Sender, amount)
	_, err = s.neutronChain.SendMsgs(mintMsg)
	s.Assert().NoError(err)

	// Send IBC transfer
	s.IBCTransfer(
		s.neutronTransferPath,
		s.neutronTransferPath.EndpointA,
		s.neutronAddr,
		s.providerAddr,
		amount.Denom,
		amount.Amount,
		"",
	)

	// Derive IBC escrow address for channel
	escrowAddress := transfertypes.GetEscrowAddress("transfer", s.neutronTransferPath.EndpointA.ChannelID)

	// Transfer tokens out of escrow address
	forceTransferMsg := tftypes.NewMsgForceTransfer(s.neutronAddr.String(), sdk.NewCoin(amount.Denom, amount.Amount), escrowAddress.String(), s.neutronAddr.String())
	_, err = s.neutronChain.SendMsgs(forceTransferMsg)
	s.Assert().ErrorContains(err, "force transfer from IBC escrow accounts is forbidden")
}

func (s *TokenfactoryTestSuite) TestForceTransferToIBCEscrow() {
	// Create token factory denom
	createDenomMsg := tftypes.NewMsgCreateDenom(s.neutronAddr.String(), "testtest")
	_, err := s.neutronChain.SendMsgs(createDenomMsg)
	s.Assert().NoError(err)

	// Derive full token factory denom
	denom := fmt.Sprintf("factory/%s/%s", createDenomMsg.Sender, createDenomMsg.Subdenom)

	// Mint denom to sender
	amount := sdk.NewCoin(denom, math.NewInt(10000000))
	mintMsg := tftypes.NewMsgMint(createDenomMsg.Sender, amount)
	_, err = s.neutronChain.SendMsgs(mintMsg)
	s.Assert().NoError(err)

	// Send IBC transfer
	s.IBCTransfer(
		s.neutronTransferPath,
		s.neutronTransferPath.EndpointA,
		s.neutronAddr,
		s.providerAddr,
		amount.Denom,
		amount.Amount,
		"",
	)

	// Derive IBC escrow address for channel
	escrowAddress := transfertypes.GetEscrowAddress("transfer", s.neutronTransferPath.EndpointA.ChannelID)

	// Transfer tokens out of escrow address
	forceTransferMsg := tftypes.NewMsgForceTransfer(s.neutronAddr.String(), sdk.NewCoin(amount.Denom, amount.Amount), s.neutronAddr.String(), escrowAddress.String())
	_, err = s.neutronChain.SendMsgs(forceTransferMsg)
	s.Assert().ErrorContains(err, "force transfer to IBC escrow accounts is forbidden")
}

func (s *TokenfactoryTestSuite) TestBurnFromIBCEscrow() {
	// Create token factory denom
	createDenomMsg := tftypes.NewMsgCreateDenom(s.neutronAddr.String(), "testtest")
	_, err := s.neutronChain.SendMsgs(createDenomMsg)
	s.Assert().NoError(err)

	// Derive full token factory denom
	denom := fmt.Sprintf("factory/%s/%s", createDenomMsg.Sender, createDenomMsg.Subdenom)

	// Mint denom to sender
	amount := sdk.NewCoin(denom, math.NewInt(10000000))
	mintMsg := tftypes.NewMsgMint(createDenomMsg.Sender, amount)
	_, err = s.neutronChain.SendMsgs(mintMsg)
	s.Assert().NoError(err)

	// Send IBC transfer
	s.IBCTransfer(
		s.neutronTransferPath,
		s.neutronTransferPath.EndpointA,
		s.neutronAddr,
		s.providerAddr,
		amount.Denom,
		amount.Amount,
		"",
	)

	// Derive IBC escrow address for channel
	escrowAddress := transfertypes.GetEscrowAddress("transfer", s.neutronTransferPath.EndpointA.ChannelID)

	// Burn tokens from escrow address
	burnMsg := tftypes.NewMsgBurnFrom(s.neutronAddr.String(), amount, escrowAddress.String())
	_, err = s.neutronChain.SendMsgs(burnMsg)
	s.Assert().ErrorContains(err, "burning from IBC escrow accounts is forbidden")
}

func (s *TokenfactoryTestSuite) TestBurnFromIBCEscrowReverse() {
	// Create token factory denom
	createDenomMsg := tftypes.NewMsgCreateDenom(s.neutronAddr.String(), "testtest")
	_, err := s.neutronChain.SendMsgs(createDenomMsg)
	s.Assert().NoError(err)

	// Derive full token factory denom
	denom := fmt.Sprintf("factory/%s/%s", createDenomMsg.Sender, createDenomMsg.Subdenom)

	// Mint denom to sender
	amount := sdk.NewCoin(denom, math.NewInt(10000000))
	mintMsg := tftypes.NewMsgMint(createDenomMsg.Sender, amount)
	_, err = s.neutronChain.SendMsgs(mintMsg)
	s.Assert().NoError(err)

	// Send IBC transfer
	s.IBCTransfer(
		s.neutronChainBPath,
		s.neutronChainBPath.EndpointA,
		s.neutronAddr,
		s.neutronAddr,
		amount.Denom,
		amount.Amount,
		"",
	)

	// Create token factory denom
	createDenomMsg = tftypes.NewMsgCreateDenom(s.bundleB.Chain.SenderAccount.GetAddress().String(), "testtest")
	_, err = s.bundleB.Chain.SendMsgs(createDenomMsg)
	s.Assert().NoError(err)

	// Derive IBC escrow address for channel
	escrowAddress := transfertypes.GetEscrowAddress("transfer", s.neutronChainBPath.EndpointB.ChannelID)

	// Derive full token factory denom
	denom = fmt.Sprintf("factory/%s/%s", s.bundleB.Chain.SenderAccount.GetAddress().String(), createDenomMsg.Subdenom)

	// Mint denom to sender
	amount = sdk.NewCoin(denom, math.NewInt(10000000))

	// Burn tokens from escrow address
	burnMsg := tftypes.NewMsgBurnFrom(s.bundleB.Chain.SenderAccount.GetAddress().String(), amount, escrowAddress.String())
	_, err = s.bundleB.Chain.SendMsgs(burnMsg)
	s.Assert().ErrorContains(err, "burning from IBC escrow accounts is forbidden")
}
