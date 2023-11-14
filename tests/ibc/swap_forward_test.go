package ibc_test

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/iancoleman/orderedmap"
	"golang.org/x/exp/maps"

	"github.com/neutron-org/neutron/x/dex/types"
	swaptypes "github.com/neutron-org/neutron/x/ibcswap/types"
)

func (s *IBCTestSuite) TestSwapAndForward_Success() {
	// Send an IBC transfer from provider chain to neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

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

	// Assert that the deposit was successful and the funds are moved out of the Neutron user acc
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())
	postDepositNeutronBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative)

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

	bz, err := json.Marshal(forwardMetadata)
	s.Assert().NoError(err)

	nextJSON := new(swaptypes.JSONObject)
	err = json.Unmarshal(bz, nextJSON)
	s.Assert().NoError(err)

	metadata := swaptypes.PacketMetadata{
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
			Next: nextJSON,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send an IBC transfer from provider to neutron with packet memo containing the swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Relay the packets
	err = s.RelayAllPacketsAToB(s.neutronChainBPath)
	s.Assert().NoError(err)

	// Check that the funds are moved out of the acc on providerChain
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)

	// Check that the amountIn is deducted from the neutron overrid receiver account
	overrideAddr := s.ReceiverOverrideAddr(s.neutronTransferPath.EndpointA.ChannelID, s.providerAddr.String())
	s.assertNeutronBalance(overrideAddr, s.providerToNeutronDenom, math.OneInt())
	// Check that neutron account did not keep any of the transfer denom
	s.assertNeutronBalance(overrideAddr, nativeDenom, math.ZeroInt())

	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronChainBPath.EndpointA.ChannelID,
		nativeDenom,
	)
	transferDenomNeutronChainB := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()

	// Check that the funds are now present in the acc on chainB
	s.assertChainBBalance(chainBAddr, transferDenomNeutronChainB, expectedAmountOut)
}

func (s *IBCTestSuite) TestSwapAndForward_MultiHopSuccess() {
	// Send an IBC transfer from provider chain to neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

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

	// Assert that the deposit was successful and the funds are moved out of the Neutron user acc
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())
	postDepositNeutronBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)

	expectedOut := math.NewInt(99_990)

	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()
	chainCAddr := s.bundleC.Chain.SenderAccount.GetAddress()

	retries := uint8(0)
	nextForward := pfmtypes.PacketMetadata{
		Forward: &pfmtypes.ForwardMetadata{
			Receiver: chainCAddr.String(),
			Port:     s.chainBChainCPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.chainBChainCPath.EndpointA.ChannelID,
			Timeout:  pfmtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nil,
		},
	}
	nextForwardBz, err := json.Marshal(nextForward)
	s.Assert().NoError(err)
	nextForwardJSON := pfmtypes.NewJSONObject(false, nextForwardBz, orderedmap.OrderedMap{})

	forwardMetadata := pfmtypes.PacketMetadata{
		Forward: &pfmtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.neutronChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.neutronChainBPath.EndpointA.ChannelID,
			Timeout:  pfmtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nextForwardJSON,
		},
	}
	bz, err := json.Marshal(forwardMetadata)
	s.Assert().NoError(err)

	nextJSON := new(swaptypes.JSONObject)
	err = json.Unmarshal(bz, nextJSON)
	s.Assert().NoError(err)

	metadata := swaptypes.PacketMetadata{
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
			Next: nextJSON,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Assert().NoError(err)

	// Send an IBC transfer from provider to neutron with packet memo containing the swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	neutronPacket := maps.Values(s.neutronChain.SentPackets)[0]
	err = s.neutronChainBPath.EndpointB.UpdateClient()
	s.Require().NoError(err)
	err = s.neutronChainBPath.EndpointB.RecvPacket(neutronPacket)
	s.Require().NoError(err)
	err = s.RelayAllPacketsAToB(s.chainBChainCPath)
	s.Require().NoError(err)

	transferDenomPathNeutronChainB := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronChainBPath.EndpointB.ChannelID,
		nativeDenom,
	)
	transferDenomNeutronChainB := transfertypes.ParseDenomTrace(transferDenomPathNeutronChainB).IBCDenom()
	transferDenomPathChainC := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.chainBChainCPath.EndpointB.ChannelID,
		transferDenomPathNeutronChainB,
	)
	transferDenomChainC := transfertypes.ParseDenomTrace(transferDenomPathChainC).IBCDenom()

	// Check that the funds are moved out of the acc on chainA
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)
	// Check that chain B balance is unchanged
	s.assertChainBBalance(chainBAddr, transferDenomNeutronChainB, math.ZeroInt())

	// Check that funds made it to chainC
	s.assertChainCBalance(chainCAddr, transferDenomChainC, expectedOut)
}

// TestSwapAndForward_UnwindIBCDenomSuccess asserts that the swap and forward middleware stack works as intended in the
// case that a native token from ChainB is sent to ChainA and then ChainA initiates a swap and forward with the token.
// This asserts that denom unwinding works as intended when going provider->neutron->provider
func (s *IBCTestSuite) TestSwapAndForward_UnwindIBCDenomSuccess() {
	// Send an IBC transfer from provider chain to neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

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
		1,
		1,
		s.neutronAddr)

	// Assert that the deposit was successful and the funds are moved out of the Neutron user acc
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())
	postDepositNeutronBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative)

	swapAmount := math.NewInt(100_000)
	expectedAmountOut := math.NewInt(99980)

	retries := uint8(0)

	forwardMetadata := pfmtypes.PacketMetadata{
		Forward: &pfmtypes.ForwardMetadata{
			Receiver: s.providerAddr.String(),
			Port:     s.neutronTransferPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.neutronTransferPath.EndpointA.ChannelID,
			Timeout:  pfmtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nil,
		},
	}

	bz, err := json.Marshal(forwardMetadata)
	s.Assert().NoError(err)

	nextJSON := new(swaptypes.JSONObject)
	err = json.Unmarshal(bz, nextJSON)
	s.Assert().NoError(err)

	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          s.neutronAddr.String(),
				Receiver:         s.neutronAddr.String(),
				TokenIn:          nativeDenom,
				TokenOut:         s.providerToNeutronDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 2,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nextJSON,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Transfer native denom from neutron to provider
	s.IBCTransfer(
		s.neutronTransferPath,
		s.neutronTransferPath.EndpointA,
		s.neutronAddr,
		s.providerAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)
	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronTransferPath.EndpointB.ChannelID,
		nativeDenom,
	)
	transferDenomNeutronProvider := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()
	s.assertProviderBalance(s.providerAddr, transferDenomNeutronProvider, ibcTransferAmount)

	// Send an IBC transfer from provider to neutron with packet memo containing the swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		transferDenomNeutronProvider,
		swapAmount,
		string(metadataBz),
	)

	// Relay the packets
	err = s.RelayAllPacketsAToB(s.neutronTransferPath)
	s.Assert().NoError(err)
	s.coordinator.CommitBlock(s.neutronChain)

	// Check that the amountIn is deducted from the neutron override receiever  account
	overrideAddr := s.ReceiverOverrideAddr(s.neutronTransferPath.EndpointA.ChannelID, s.providerAddr.String())
	s.assertNeutronBalance(overrideAddr, nativeDenom, math.OneInt())
	// Check that the funds are now present on the provider chainer
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Add(expectedAmountOut),
	)
}

// TestSwapAndForward_ForwardFailsRefundAddr asserts that the swap and forward middleware stack works as intended in the case
// that an incoming IBC swap succeeds but the forward fails when a NeutronRefundAddress is provided.
// The swap will be reverted and the transferred amount will be credited to the refundAddr
func (s *IBCTestSuite) TestSwapAndForward_ForwardFailsNoIBCRefund() {
	// Send an IBC transfer from provider chain to neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

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

	// Assert that the deposit was successful and the funds are moved out of the Neutron user acc
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())
	postDepositNeutronBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)
	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()

	retries := uint8(0)

	forwardMetadata := pfmtypes.PacketMetadata{
		Forward: &pfmtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.neutronChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  "invalid-channel", // add an invalid channel identifier so the forward fails
			Timeout:  pfmtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nil,
		},
	}

	bz, err := json.Marshal(forwardMetadata)
	s.Assert().NoError(err)

	nextJSON := new(swaptypes.JSONObject)
	err = json.Unmarshal(bz, nextJSON)
	s.Assert().NoError(err)

	metadata := swaptypes.PacketMetadata{
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
			NoIBCRefund: true,
			Next:        nextJSON,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send an IBC transfer from provider to neutron with packet memo containing the swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Relay the packets from neutron => ChainB
	err = s.RelayAllPacketsAToB(s.neutronChainBPath)
	// Relay Fails
	s.Assert().Error(err)

	// Check that the funds are moved out of the acc on providerChain
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)

	// Check that nothing remains in the overrideReceiver account
	overrideAddr := s.ReceiverOverrideAddr(s.neutronTransferPath.EndpointA.ChannelID, s.providerAddr.String())
	s.assertNeutronBalance(overrideAddr, s.providerToNeutronDenom, math.ZeroInt())
	s.assertNeutronBalance(overrideAddr, nativeDenom, math.ZeroInt())

	// Check that the swap was reverted and the transfer amount is in the creator account
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, ibcTransferAmount)

	// Check that nothing made it to chainB
	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronChainBPath.EndpointA.ChannelID,
		nativeDenom,
	)
	transferDenomNeutronChainB := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()

	s.assertChainBBalance(chainBAddr, transferDenomNeutronChainB, math.ZeroInt())
}

// TestSwapAndForward_ForwardFailsRefundAddr asserts that the swap and forward middleware stack works as intended in the case
// that an incoming IBC swap succeeds but the forward fails when no NeutronRefundAddress is provided.
// The swap will be reverted and a refund to the src chain will take place.
func (s *IBCTestSuite) TestSwapAndForward_ForwardFails() {
	// Send an IBC transfer from provider chain to neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

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

	// Assert that the deposit was successful and the funds are moved out of the Neutron user acc
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())
	postDepositNeutronBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)
	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()

	retries := uint8(0)

	forwardMetadata := pfmtypes.PacketMetadata{
		Forward: &pfmtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.neutronChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  "invalid-channel", // add an invalid channel identifier so the forward fails
			Timeout:  pfmtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nil,
		},
	}

	bz, err := json.Marshal(forwardMetadata)
	s.Assert().NoError(err)

	nextJSON := new(swaptypes.JSONObject)
	err = json.Unmarshal(bz, nextJSON)
	s.Assert().NoError(err)

	metadata := swaptypes.PacketMetadata{
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
			Next: nextJSON,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send an IBC transfer from provider to neutron with packet memo containing the swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Relay the packets from neutron => ChainB
	err = s.RelayAllPacketsAToB(s.neutronChainBPath)
	// Relay Fails
	s.Assert().Error(err)

	// Check that nothing remains in the overrideReceiver account
	overrideAddr := s.ReceiverOverrideAddr(s.neutronTransferPath.EndpointA.ChannelID, s.providerAddr.String())
	s.assertNeutronBalance(overrideAddr, s.providerToNeutronDenom, math.ZeroInt())
	s.assertNeutronBalance(overrideAddr, nativeDenom, math.ZeroInt())

	// Check that nothing made it to chainB
	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.neutronChainBPath.EndpointA.ChannelID,
		nativeDenom,
	)
	transferDenomNeutronChainB := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()

	s.assertChainBBalance(chainBAddr, transferDenomNeutronChainB, math.ZeroInt())

	// Check that the refund takes place and the funds are moved back to the account on Gaia
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount.Sub(depositAmount))
}
