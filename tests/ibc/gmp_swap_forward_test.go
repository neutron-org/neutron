package ibc_test

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/gmp"
	swaptypes "github.com/neutron-org/neutron/x/ibcswap/types"
)

// TestSwapAndForward_Success asserts that the swap and forward middleware stack works as intended with Neutron running as a
// consumer chain connected to two other chains via IBC.
func (s *IBCTestSuite) TestGMPSwapAndForward_Success() {
	// Send an IBC transfer from provider to Neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(s.providerAddr, s.neutronAddr, nativeDenom, ibcTransferAmount, "")

	// Assert that the funds are gone from the acc on provider and present in the acc on Neutron
	newProviderBalNative := genesisWalletAmount.Sub(ibcTransferAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative)

	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, ibcTransferAmount)

	// deposit stake<>ibcTransferToken to initialize the pool on Neutron
	depositAmount := math.NewInt(100_000)
	s.neutronDeposit(
		nativeDenom,
		s.providerToNeutronDenom,
		depositAmount,
		depositAmount,
		0,
		1,
		s.neutronAddr)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)
	expectedAmountOut := math.NewInt(99990)
	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()

	retries := uint8(0)
	forwardMetadata := pfmtypes.PacketMetadata{
		Forward: &pfmtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.neutronChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.neutronChainBPath.EndpointA.ChannelID,
			Timeout:  pfmtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nil,
		},
	}

	forwardBz, err := json.Marshal(forwardMetadata)
	s.Require().NoError(err)

	forwardNextJSON := new(swaptypes.JSONObject)
	err = json.Unmarshal(forwardBz, forwardNextJSON)
	s.Require().NoError(err)

	swapMetadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          s.neutronAddr.String(),
				Receiver:         s.neutronAddr.String(),
				TokenIn:          s.providerToNeutronDenom,
				TokenOut:         nativeDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 2,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: forwardNextJSON,
		},
	}
	swapMetadataBz, err := json.Marshal(swapMetadata)

	s.Require().NoError(err)

	gmpMetadata := gmp.Message{
		SourceChain:   "axelar",
		SourceAddress: "alice",
		Payload:       swapMetadataBz,
		Type:          gmp.TypeGeneralMessageWithToken,
	}

	gmpMetadataBz, err := json.Marshal(gmpMetadata)
	s.Require().NoError(err)

	// Send an IBC transfer from chainA to chainB with packet memo containing the swap metadata

	s.IBCTransferProviderToNeutron(s.providerAddr, s.neutronAddr, nativeDenom, ibcTransferAmount, string(gmpMetadataBz))

	// Relay the packet
	err = s.RelayAllPacketsAToB(s.neutronChainBPath)
	s.Assert().NoError(err)

	// Check that the funds are moved out of the acc on providerChain
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative.Sub(ibcTransferAmount))

	// Check that the amountIn is deduced from the neutron account
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())
	// Check that neutron account did not keep any of the transfer denom
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, genesisWalletAmount.Sub(swapAmount))

	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronChainBPath.EndpointA.ChannelID,
		nativeDenom,
	)
	transferDenomChainB := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()

	// Check that the funds are now present in the acc on chainB
	s.assertChainBBalance(chainBAddr, transferDenomChainB, expectedAmountOut)

	s.Assert().NoError(err)
}
