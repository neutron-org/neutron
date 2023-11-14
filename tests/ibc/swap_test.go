package ibc_test

import (
	"encoding/json"

	"cosmossdk.io/math"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	swaptypes "github.com/neutron-org/neutron/x/ibcswap/types"
)

// TestIBCSwapMiddleware_Success asserts that the IBC swap middleware works as intended with Neutron running as a
// consumer chain connected to the Cosmos Hub.
func (s *IBCTestSuite) TestIBCSwapMiddleware_Success() {
	// Send an IBC transfer from provider to Neutron, so we can initialize a pool with the IBC denom token + native Neutron token
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

	// Send an IBC transfer from providerChain to Neutron with packet memo containing the swap metadata
	swapAmount := math.NewInt(100000)
	expectedOut := math.NewInt(99_990)

	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.neutronAddr.String(),
				Receiver:         s.neutronAddr.String(),
				TokenIn:          s.providerToNeutronDenom,
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

	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
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

	// Check that the swap funds are now present in the acc on Neutron
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative.Add(expectedOut))

	// Check that all of the IBC transfer denom have been used up
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.OneInt())
}

// TestIBCSwapMiddleware_FailRefund asserts that the IBC swap middleware works as intended with Neutron running as a
// consumer chain connected to the Cosmos Hub. The swap should fail and a refund to the src chain should take place.
func (s *IBCTestSuite) TestIBCSwapMiddleware_FailRefund() {
	// Compose the swap metadata, this swap will fail because there is no pool initialized for this pair
	swapAmount := math.NewInt(100000)
	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.neutronAddr.String(),
				Receiver:         s.neutronAddr.String(),
				TokenIn:          s.providerToNeutronDenom,
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

	// Send (failing) IBC transfer with  swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds are not present in the account on Neutron
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, genesisWalletAmount)
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())

	// Check that the refund takes place and the funds are moved back to the account on Gaia
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount)
}

func (s *IBCTestSuite) TestIBCSwapMiddleware_FailWithRefundAddr() {
	// Compose the swap metadata, this swap will fail because there is no pool initialized for this pair
	refundAddr := s.neutronChain.SenderAccounts[1].SenderAccount.GetAddress()
	swapAmount := math.NewInt(100000)
	metadata := swaptypes.PacketMetadata{
		Swap: &swaptypes.SwapMetadata{
			MsgPlaceLimitOrder: &dextypes.MsgPlaceLimitOrder{
				Creator:          s.neutronAddr.String(),
				Receiver:         s.neutronAddr.String(),
				TokenIn:          s.providerToNeutronDenom,
				TokenOut:         nativeDenom,
				AmountIn:         swapAmount,
				TickIndexInToOut: 1,
				OrderType:        dextypes.LimitOrderType_FILL_OR_KILL,
			},
			NeutronRefundAddress: refundAddr.String(),
			Next:                 nil,
		},
	}

	metadataBz, err := json.Marshal(metadata)
	s.Require().NoError(err)

	// Send (failing) IBC transfer with swap metadata
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds have been moved to the refund address
	s.assertNeutronBalance(refundAddr, nativeDenom, genesisWalletAmount)
	s.assertNeutronBalance(refundAddr, s.providerToNeutronDenom, ibcTransferAmount)

	// Check that no refund takes place and the funds are not in the account on provider
	s.assertProviderBalance(s.providerAddr, nativeDenom, genesisWalletAmount.Sub(ibcTransferAmount))
}
