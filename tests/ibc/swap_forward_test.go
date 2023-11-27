package ibc_test

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	"github.com/neutron-org/neutron/x/dex/types"
	swaptypes "github.com/neutron-org/neutron/x/ibcswap/types"
)

// TestSwapAndForward_Fails asserts that the IBC swap middleware fails gracefully when provided with a package-forward memo
func (s *IBCTestSuite) TestSwapAndForward_Fails() {
	// Send an IBC transfer from provider chain to neutron, so we can initialize a pool with the IBC denom token + native Neutron token
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		"",
	)
	newProviderBalNative := genesisWalletAmount.Sub(ibcTransferAmount)

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

	postDepositNeutronBalNative := genesisWalletAmount.Sub(depositAmount)

	// Compose the IBC transfer memo metadata to be used in the swap and forward
	swapAmount := math.NewInt(100000)
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

	// Send (failing) IBC transfer with PFM data
	s.IBCTransferProviderToNeutron(
		s.providerAddr,
		s.neutronAddr,
		nativeDenom,
		ibcTransferAmount,
		string(metadataBz),
	)

	// Check that the funds are not present in the account on Neutron
	s.assertNeutronBalance(s.neutronAddr, nativeDenom, postDepositNeutronBalNative)
	s.assertNeutronBalance(s.neutronAddr, s.providerToNeutronDenom, math.ZeroInt())

	// Check that the refund takes place and the funds are moved back to the account on Gaia
	s.assertProviderBalance(s.providerAddr, nativeDenom, newProviderBalNative)
}
