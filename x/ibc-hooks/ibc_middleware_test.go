package ibchooks_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/abci/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/app/params"
	"github.com/neutron-org/neutron/v6/testutil"
	"github.com/neutron-org/neutron/v6/x/ibc-hooks/testutils"
	"github.com/neutron-org/neutron/v6/x/ibc-hooks/utils"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

type HooksTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestIBCHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func (suite *HooksTestSuite) TestOnRecvPacketHooks() {
	var (
		trace    transfertypes.DenomTrace
		amount   math.Int
		receiver string
		status   testutils.Status
	)

	testCases := []struct {
		msg      string
		malleate func(*testutils.Status)
		expPass  bool
	}{
		{"override", func(status *testutils.Status) {
			suite.GetNeutronZoneApp(suite.ChainB).TransferStack.
				ICS4Middleware.Hooks = testutils.TestRecvOverrideHooks{Status: status}
		}, true},
		{"before and after", func(status *testutils.Status) {
			suite.GetNeutronZoneApp(suite.ChainB).TransferStack.
				ICS4Middleware.Hooks = testutils.TestRecvBeforeAfterHooks{Status: status}
		}, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			suite.ConfigureTransferChannel()
			receiver = suite.ChainB.SenderAccount.GetAddress().String() // must be explicitly changed in malleate
			status = testutils.Status{}

			amount = math.NewInt(100) // must be explicitly changed in malleate
			seq := uint64(1)

			trace = transfertypes.ParseDenomTrace(params.DefaultDenom)

			// send coin from chainA to chainB
			transferMsg := transfertypes.NewMsgTransfer(
				suite.TransferPath.EndpointA.ChannelConfig.PortID,
				suite.TransferPath.EndpointA.ChannelID,
				sdk.NewCoin(trace.IBCDenom(), amount),
				suite.ChainA.SenderAccount.GetAddress().String(),
				receiver,
				clienttypes.NewHeight(1, 110),
				0, "")
			_, err := suite.ChainA.SendMsgs(transferMsg)
			suite.Require().NoError(err) // message committed

			tc.malleate(&status)

			data := transfertypes.NewFungibleTokenPacketData(trace.GetFullDenomPath(), amount.String(), suite.ChainA.SenderAccount.GetAddress().String(), receiver, "")
			packet := channeltypes.NewPacket(data.GetBytes(), seq, suite.TransferPath.EndpointA.ChannelConfig.PortID, suite.TransferPath.EndpointA.ChannelID, suite.TransferPath.EndpointB.ChannelConfig.PortID, suite.TransferPath.EndpointB.ChannelID, clienttypes.NewHeight(1, 100), 0)

			ack := suite.GetNeutronZoneApp(suite.ChainB).TransferStack.
				OnRecvPacket(suite.ChainB.GetContext(), packet, suite.ChainA.SenderAccount.GetAddress())

			if tc.expPass {
				suite.Require().True(ack.Success())
			} else {
				suite.Require().False(ack.Success())
			}

			if _, ok := suite.GetNeutronZoneApp(suite.ChainB).TransferStack.
				ICS4Middleware.Hooks.(testutils.TestRecvOverrideHooks); ok {
				suite.Require().True(status.OverrideRan)
				suite.Require().False(status.BeforeRan)
				suite.Require().False(status.AfterRan)
			}

			if _, ok := suite.GetNeutronZoneApp(suite.ChainB).TransferStack.
				ICS4Middleware.Hooks.(testutils.TestRecvBeforeAfterHooks); ok {
				suite.Require().False(status.OverrideRan)
				suite.Require().True(status.BeforeRan)
				suite.Require().True(status.AfterRan)
			}
		})
	}
}

func (suite *HooksTestSuite) makeMockPacket(receiver, memo string, prevSequence uint64) channeltypes.Packet {
	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    sdk.DefaultBondDenom,
		Amount:   "1",
		Sender:   suite.ChainB.SenderAccount.GetAddress().String(),
		Receiver: receiver,
		Memo:     memo,
	}

	return channeltypes.NewPacket(
		packetData.GetBytes(),
		prevSequence+1,
		suite.TransferPath.EndpointB.ChannelConfig.PortID,
		suite.TransferPath.EndpointB.ChannelID,
		suite.TransferPath.EndpointA.ChannelConfig.PortID,
		suite.TransferPath.EndpointA.ChannelID,
		clienttypes.NewHeight(0, 150),
		0,
	)
}

func (suite *HooksTestSuite) receivePacket(receiver, memo string) []byte {
	return suite.receivePacketWithSequence(receiver, memo, 0)
}

func (suite *HooksTestSuite) receivePacketWithSequence(receiver, memo string, prevSequence uint64) []byte {
	channelCap := suite.ChainB.GetChannelCapability(
		suite.TransferPath.EndpointB.ChannelConfig.PortID,
		suite.TransferPath.EndpointB.ChannelID)

	packet := suite.makeMockPacket(receiver, memo, prevSequence)

	seqID, err := suite.GetNeutronZoneApp(suite.ChainB).HooksICS4Wrapper.SendPacket(
		suite.ChainB.GetContext(), channelCap, suite.TransferPath.EndpointA.ChannelConfig.PortID, suite.TransferPath.EndpointA.ChannelID, packet.TimeoutHeight, packet.TimeoutTimestamp, packet.Data)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)
	suite.Require().Equal(prevSequence+1, seqID)

	// Update both clients
	err = suite.TransferPath.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.TransferPath.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// recv in chain a
	res, err := suite.TransferPath.EndpointA.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	// get the ack from the chain a's response
	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// manually send the acknowledgement to chain b
	err = suite.TransferPath.EndpointA.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)
	return ack
}

func (suite *HooksTestSuite) TestRecvTransferWithMetadata() {
	suite.ConfigureTransferChannel()

	// Setup contract
	codeID := suite.StoreContractCode(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/echo.wasm")
	addr := suite.InstantiateContract(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeID, "{}")

	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")
}

// After successfully executing a wasm call, the contract should have the funds sent via IBC
func (suite *HooksTestSuite) TestFundsAreTransferredToTheContract() {
	suite.ConfigureTransferChannel()

	// Setup contract
	codeID := suite.StoreContractCode(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/echo.wasm")
	addr := suite.InstantiateContract(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeID, "{}")

	// Check that the contract has no funds
	localDenom := utils.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
	balance := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), addr, localDenom)
	suite.Require().Equal(math.NewInt(0), balance.Amount)

	// Execute the contract via IBC
	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")

	// Check that the token has now been transferred to the contract
	balance = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), addr, localDenom)
	suite.Require().Equal(math.NewInt(1), balance.Amount)
}

// If the wasm call wails, the contract acknowledgement should be an error and the funds returned
func (suite *HooksTestSuite) TestFundsAreReturnedOnFailedContractExec() {
	suite.ConfigureTransferChannel()

	// Setup contract
	codeID := suite.StoreContractCode(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/echo.wasm")
	addr := suite.InstantiateContract(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeID, "{}")

	// Check that the contract has no funds
	localDenom := utils.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
	balance := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), addr, localDenom)
	suite.Require().Equal(math.NewInt(0), balance.Amount)

	// Execute the contract via IBC with a message that the contract will reject
	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"not_echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().Contains(ack, "error")

	// Check that the token has now been transferred to the contract
	balance = suite.GetNeutronZoneApp(suite.ChainA).BankKeeper.GetBalance(suite.ChainA.GetContext(), addr, localDenom)
	fmt.Println(balance)
	suite.Require().Equal(math.NewInt(0), balance.Amount)
}

func (suite *HooksTestSuite) TestPacketsThatShouldBeSkipped() {
	suite.ConfigureTransferChannel()

	var sequence uint64
	receiver := suite.ChainB.SenderAccount.GetAddress().String()

	testCases := []struct {
		memo           string
		expPassthrough bool
	}{
		{"", true},
		{"{01]", true}, // bad json
		{"{}", true},
		{`{"something": ""}`, true},
		{`{"wasm": "test"}`, false},
		{`{"wasm": []`, true}, // invalid top level JSON
		{`{"wasm": {}`, true}, // invalid top level JSON
		{`{"wasm": []}`, false},
		{`{"wasm": {}}`, false},
		{`{"wasm": {"contract": "something"}}`, false},
		{`{"wasm": {"contract": "cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj"}}`, false},
		{`{"wasm": {"msg": "something"}}`, false},
		// invalid receiver
		{`{"wasm": {"contract": "cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj", "msg": {}}}`, false},
		// msg not an object
		{fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": 1}}`, receiver), false},
	}

	for _, tc := range testCases {
		ackBytes := suite.receivePacketWithSequence(receiver, tc.memo, sequence)
		ackStr := string(ackBytes)
		// fmt.Println(ackStr)
		var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
		err := json.Unmarshal(ackBytes, &ack)
		suite.Require().NoError(err)
		if tc.expPassthrough {
			suite.Require().Equal("AQ==", ack["result"], tc.memo)
		} else {
			suite.Require().Contains(ackStr, "error", tc.memo)
		}
		sequence++
	}
}

type Direction int64

const (
	AtoB Direction = iota
	BtoA
)

func (suite *HooksTestSuite) GetEndpoints(direction Direction) (sender, receiver *ibctesting.Endpoint) {
	switch direction {
	case AtoB:
		sender = suite.TransferPath.EndpointA
		receiver = suite.TransferPath.EndpointB
	case BtoA:
		sender = suite.TransferPath.EndpointB
		receiver = suite.TransferPath.EndpointA
	}
	return sender, receiver
}

func (suite *HooksTestSuite) RelayPacket(packet channeltypes.Packet, direction Direction) (*types.ExecTxResult, []byte) {
	sender, receiver := suite.GetEndpoints(direction)

	err := receiver.UpdateClient()
	suite.Require().NoError(err)

	// receiver Receives
	receiveResult, err := receiver.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)

	// sender Acknowledges
	err = sender.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)

	err = sender.UpdateClient()
	suite.Require().NoError(err)
	err = receiver.UpdateClient()
	suite.Require().NoError(err)

	return receiveResult, ack
}

func (suite *HooksTestSuite) StoreContractCode(chain *ibctesting.TestChain, addr sdk.AccAddress, path string) uint64 {
	wasmCode, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	codeID, _, err := wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(chain).WasmKeeper).Create(chain.GetContext(), addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody})
	if err != nil {
		panic(err)
	}

	return codeID
}

func (suite *HooksTestSuite) InstantiateContract(chain *ibctesting.TestChain, funder sdk.AccAddress, codeID uint64, initMsg string) sdk.AccAddress {
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(chain).WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(chain.GetContext(), codeID, funder, funder, []byte(initMsg), "demo contract", nil)
	if err != nil {
		panic(err)
	}
	return addr
}

func (suite *HooksTestSuite) QueryContract(chain *ibctesting.TestChain, contract sdk.AccAddress, req []byte) string {
	state, err := suite.GetNeutronZoneApp(chain).WasmKeeper.QuerySmart(chain.GetContext(), contract, req)
	if err != nil {
		panic(err)
	}
	return string(state)
}
