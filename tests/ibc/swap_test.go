package ibc_test

import (
	"encoding/json"

	"cosmossdk.io/math"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	swaptypes "github.com/neutron-org/neutron/x/ibcswap/types"
)

// TestIBCSwapMiddleware_Success asserts that the IBC swap middleware works as intended with Duality running as a
// consumer chain connected to the Cosmos Hub.
func (s *IBCTestSuite) TestIBCSwapMiddleware_Success() {
	// Send an IBC transfer from provider to Duality, so we can initialize a pool with the IBC denom token + native Duality token
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

	// Send an IBC transfer from providerChain to Duality with packet memo containing the swap metadata
	swapAmount := math.NewInt(100000)
	expectedOut := math.NewInt(99_990)

	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
				TokenOut:         nativeDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 1,
				OrderType:        dextypes.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds are moved out of the acc on providerChain
	s.assertProviderBalance(
		s.providerAddr,
		nativeDenom,
		newProviderBalNative.Sub(ibcTransferAmount),
	)

	// Check that the swap funds are now present in the acc on Duality
	s.assertDualityBalance(s.dualityAddr, nativeDenom, postDepositDualityBalNative.Add(expectedOut))

	// Check that all of the IBC transfer denom have been used up
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())
}

// TestIBCSwapMiddleware_FailRefund asserts that the IBC swap middleware works as intended with Duality running as a
// consumer chain connected to the Cosmos Hub. The swap should fail and a refund to the src chain should take place.
func (s *IBCTestSuite) TestIBCSwapMiddleware_FailRefund() {
	// Compose the swap metadata, this swap will fail because there is no pool initialized for this pair
	swapAmount := math.NewInt(100000)
	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
				TokenOut:         nativeDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 1,
				OrderType:        dextypes.LimitOrderType_FILL_OR_KILL,
			},
			NonRefundable: false,
			Next:          nil,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send (failing) IBC transfer with  swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds are not present in the account on Duality
	s.assertDualityBalance(s.dualityAddr, nativeDenom, genesisWalletAmount)
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, math.ZeroInt())

	// Check that the refund takes place and the funds are moved back to the account on Gaia
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount)
}

// TestIBCSwapMiddleware_FailNoRefund asserts that the IBC swap middleware works as intended with Duality running as a
// consumer chain connected to the Cosmos Hub. The swap should fail and funds should remain on Duality.
func (s *IBCTestSuite) TestIBCSwapMiddleware_FailNoRefund() {
	// Compose the swap metadata, this swap will fail because there is no pool initialized for this pair
	swapAmount := math.NewInt(100000)
	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
				TokenOut:         nativeDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 1,
				OrderType:        dextypes.LimitOrderType_FILL_OR_KILL,
			},
			NonRefundable: true,
			Next:          nil,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send (failing) IBC transfer with swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds are present in the account on Duality
	s.assertDualityBalance(s.dualityAddr, nativeDenom, genesisWalletAmount)
	s.assertDualityBalance(s.dualityAddr, s.providerToDualityDenom, ibcTransferAmount)

	// Check that no refund takes place and the funds are not in the account on provider
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount.Sub(ibcTransferAmount))
}

// TestIBCSwapMiddleware_FailWithRefundAddr asserts that the IBC swap middleware works as intended with Duality running as a
// consumer chain connected to the Cosmos Hub. The swap should fail and funds should remain on Duality but be moved
// to the refund address.

func (s *IBCTestSuite) TestIBCSwapMiddleware_FailWithRefundAddr() {
	// Compose the swap metadata, this swap will fail because there is no pool initialized for this pair
	refundAddr := s.dualityChain.SenderAccounts[1].SenderAccount.GetAddress()
	swapAmount := math.NewInt(100000)
	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.dualityAddr.String(),
				Receiver:         s.dualityAddr.String(),
				TokenIn:          s.providerToDualityDenom,
				TokenOut:         nativeDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 1,
				OrderType:        dextypes.LimitOrderType_FILL_OR_KILL,
			},
			RefundAddress: refundAddr.String(),
			NonRefundable: true,
			Next:          nil,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send (failing) IBC transfer with swap metadata
	s.IBCTransferProviderToDuality(
		s.providerAddr,
		s.dualityAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds have been moved to the refund address
	s.assertDualityBalance(refundAddr, nativeDenom, genesisWalletAmount)
	s.assertDualityBalance(refundAddr, s.providerToDualityDenom, ibcTransferAmount)

	// Check that no refund takes place and the funds are not in the account on provider
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount.Sub(ibcTransferAmount))
}
