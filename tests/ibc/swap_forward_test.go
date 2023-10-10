package ibc_test

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	forwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/iancoleman/orderedmap"
	"github.com/neutron-org/neutron/x/dex/types"
	swaptypes "github.com/neutron-org/neutron/x/ibcswap/types"
	"golang.org/x/exp/maps"
)

func (s *IBCTestSuite) TestSwapAndForward_Success() {
	// Send an IBC transfer from provider chain to duality, so we can initialize a pool with the IBC denom token + native Duality token
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

	// Assert that the funds are gone from the acc on provider and present in the acc on Duality
	newProviderBalNative := genesisWalletAmount.Sub(ibcTransferAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative)

	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, ibcTransferAmount)

	// deposit stake<>ibcTransferToken to initialize the pool on Duality
	depositAmount := math.NewInt(100_000)
	s.dualityDeposit(
		nativeDenom,
		s.providerToDualityDenom,
		depositAmount,
		depositAmount,
		0,
		1,
		s.dualityAddr)

	// Assert that the deposit was successful and the funds are moved out of the Duality user acc
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
	postDepositDualityBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)
	expectedAmountOut := math.NewInt(99990)
	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()

	retries := uint8(0)

	forwardMetadata := forwardtypes.PacketMetadata{
		Forward: &forwardtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.dualityChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.dualityChainBPath.EndpointA.ChannelID,
			Timeout:  forwardtypes.Duration(5 * time.Minute),
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
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
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

	// Send an IBC transfer from provider to duality with packet memo containing the swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Relay the packets
	err = s.RelayAllPacketsAToB(s.dualityChainBPath)
	s.Assert().NoError(err)

	// Check that the funds are moved out of the acc on providerChain
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)

	// Check that the amountIn is deduced from the duality account
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
	// Check that duality account did not keep any of the transfer denom
	s.assertDualityBalance(s.dualityAddr, nativeDenom, genesisWalletAmount.Sub(swapAmount))

	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.dualityChainBPath.EndpointA.ChannelID,
		nativeDenom,
	)
	transferDenomDuality_B := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()

	// Check that the funds are now present in the acc on chainB
	s.assertChainBBalance(chainBAddr, transferDenomDuality_B, expectedAmountOut)

	s.Assert().NoError(err)
}

func (s *IBCTestSuite) TestSwapAndForward_MultiHopSuccess() {
	// Send an IBC transfer from provider chain to duality, so we can initialize a pool with the IBC denom token + native Duality token
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

	// Assert that the funds are gone from the acc on provider and present in the acc on Duality
	newProviderBalNative := genesisWalletAmount.Sub(ibcTransferAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative)

	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, ibcTransferAmount)

	// deposit stake<>ibcTransferToken to initialize the pool on Duality
	depositAmount := math.NewInt(100_000)
	s.dualityDeposit(
		nativeDenom,
		s.providerToDualityDenom,
		depositAmount,
		depositAmount,
		0,
		1,
		s.dualityAddr)

	// Assert that the deposit was successful and the funds are moved out of the Duality user acc
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
	postDepositDualityBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)

	expectedOut := math.NewInt(99_990)

	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()
	chainCAddr := s.bundleC.Chain.SenderAccount.GetAddress()

	retries := uint8(0)
	nextForward := forwardtypes.PacketMetadata{
		Forward: &forwardtypes.ForwardMetadata{
			Receiver: chainCAddr.String(),
			Port:     s.chainBChainCPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.chainBChainCPath.EndpointA.ChannelID,
			Timeout:  forwardtypes.Duration(5 * time.Minute),
			Retries:  &retries,
			Next:     nil,
		},
	}
	nextForwardBz, err := json.Marshal(nextForward)
	s.Assert().NoError(err)
	nextForwardJSON := forwardtypes.NewJSONObject(false, nextForwardBz, orderedmap.OrderedMap{})

	forwardMetadata := forwardtypes.PacketMetadata{
		Forward: &forwardtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.dualityChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.dualityChainBPath.EndpointA.ChannelID,
			Timeout:  forwardtypes.Duration(5 * time.Minute),
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
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
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

	// Send an IBC transfer from provider to duality with packet memo containing the swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	dualityPacket := maps.Values(s.dualityChain.SentPackets)[0]
	err = s.dualityChainBPath.EndpointB.UpdateClient()
	s.Require().NoError(err)
	err = s.dualityChainBPath.EndpointB.RecvPacket(dualityPacket)
	s.Require().NoError(err)
	err = s.RelayAllPacketsAToB(s.chainBChainCPath)
	s.Require().NoError(err)

	transferDenomPathDuality_B := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.dualityChainBPath.EndpointB.ChannelID,
		nativeDenom,
	)
	transferDenomDuality_B := transfertypes.ParseDenomTrace(transferDenomPathDuality_B).IBCDenom()
	transferDenomPathB_C := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.chainBChainCPath.EndpointB.ChannelID,
		transferDenomPathDuality_B,
	)
	transferDenomB_C := transfertypes.ParseDenomTrace(transferDenomPathB_C).IBCDenom()

	// Check that the funds are moved out of the acc on chainA
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)
	// Check that chain B balance is unchanged
	s.assertChainBBalance(chainBAddr, transferDenomDuality_B, math.ZeroInt())

	// Check that funds made it to chainC
	s.assertChainCBalance(chainCAddr, transferDenomB_C, expectedOut)
}

// TestSwapAndForward_UnwindIBCDenomSuccess asserts that the swap and forward middleware stack works as intended in the
// case that a native token from ChainB is sent to ChainA and then ChainA initiates a swap and forward with the token.
// This asserts that denom unwinding works as intended when going provider->duality->provider
func (s *IBCTestSuite) TestSwapAndForward_UnwindIBCDenomSuccess() {
	// Send an IBC transfer from provider chain to duality, so we can initialize a pool with the IBC denom token + native Duality token
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

	// Assert that the funds are gone from the acc on provider and present in the acc on Duality
	newProviderBalNative := genesisWalletAmount.Sub(ibcTransferAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative)

	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, ibcTransferAmount)

	// deposit stake<>ibcTransferToken to initialize the pool on Duality
	depositAmount := math.NewInt(100_000)
	s.dualityDeposit(
		nativeDenom,
		s.providerToDualityDenom,
		depositAmount,
		depositAmount,
		0,
		1,
		s.dualityAddr)

	// Assert that the deposit was successful and the funds are moved out of the Duality user acc
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
	postDepositDualityBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative)

	swapAmount := math.NewInt(100000)
	expectedAmountOut := math.NewInt(99990)

	retries := uint8(0)

	forwardMetadata := forwardtypes.PacketMetadata{
		Forward: &forwardtypes.ForwardMetadata{
			Receiver: s.providerAddr.String(),
			Port:     s.dualityTransferPath.EndpointA.ChannelConfig.PortID,
			Channel:  s.dualityTransferPath.EndpointA.ChannelID,
			Timeout:  forwardtypes.Duration(5 * time.Minute),
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
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          nativeDenom,
				TokenOut:         s.providerToDualityDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 2,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nextJSON,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send an IBC transfer from provider to duality with packet memo containing the swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Relay the packets
	s.RelayAllPacketsAToB(s.dualityTransferPath)
	s.coordinator.CommitBlock(s.dualityChain)

	// Check that the amountIn is deduced from the duality account
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative.Sub(swapAmount))
	// Check that the amountIn has been deducted from the duality chain
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative.Sub(swapAmount))
	// Check that the funds are now present on the provider chainer
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount).Add(expectedAmountOut),
	)

	s.Assert().NoError(err)
}

// TestSwapAndForward_ForwardFails asserts that the swap and forward middleware stack works as intended in the case
// that an incoming IBC swap succeeds but the forward fails.
func (s *IBCTestSuite) TestSwapAndForward_ForwardFails() {
	// Send an IBC transfer from provider chain to duality, so we can initialize a pool with the IBC denom token + native Duality token
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)

	// Assert that the funds are gone from the acc on provider and present in the acc on Duality
	newProviderBalNative := genesisWalletAmount.Sub(ibcTransferAmount)
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative)

	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, ibcTransferAmount)

	// deposit stake<>ibcTransferToken to initialize the pool on Duality
	depositAmount := math.NewInt(100_000)
	s.dualityDeposit(
		nativeDenom,
		s.providerToDualityDenom,
		depositAmount,
		depositAmount,
		0,
		1,
		s.dualityAddr)

	// Assert that the deposit was successful and the funds are moved out of the Duality user acc
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
	postDepositDualityBalNative := genesisWalletAmount.Sub(depositAmount)
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)
	expectedAmountOut := math.NewInt(99990)
	chainBAddr := s.bundleB.Chain.SenderAccount.GetAddress()

	retries := uint8(0)

	forwardMetadata := forwardtypes.PacketMetadata{
		Forward: &forwardtypes.ForwardMetadata{
			Receiver: chainBAddr.String(),
			Port:     s.dualityChainBPath.EndpointA.ChannelConfig.PortID,
			Channel:  "invalid-channel", // add an invalid channel identifier so the forward fails
			Timeout:  forwardtypes.Duration(5 * time.Minute),
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
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
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

	// Send an IBC transfer from provider to duality with packet memo containing the swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Relay the packets from duality => ChainB
	err = s.RelayAllPacketsAToB(s.dualityChainBPath)
	// Relay Fails
	s.Assert().Error(err)

	// Check that the funds are moved out of the acc on providerChain
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)

	// Check that the amountIn is deduced from the duality account
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
	// Check that the amountOut stays on the dualitychain
	s.assertDualityBalance(
		s.dualityAddr,
		nativeDenom,
		postDepositDualityBalNative.Add(expectedAmountOut),
	)

	// Check that nothing made it to chainB
	transferDenomPath := transfertypes.GetPrefixedDenom(
		transfertypes.PortID,
		s.dualityChainBPath.EndpointA.ChannelID,
		nativeDenom,
	)
	transferDenomDuality_B := transfertypes.ParseDenomTrace(transferDenomPath).IBCDenom()

	s.assertChainBBalance(chainBAddr, transferDenomDuality_B, math.ZeroInt())
}
