package ibchooks_test

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	ibctesting "github.com/cosmos/interchain-security/legacy_ibc_testing/testing"
	"github.com/neutron-org/neutron/app/params"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/x/ibc-hooks/testutils"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
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
		amount   sdk.Int
		receiver string
		status   testutils.Status
	)

	testCases := []struct {
		msg      string
		malleate func(*testutils.Status)
		expPass  bool
	}{
		{"override", func(status *testutils.Status) {
			suite.GetNeutronZoneApp(suite.ChainB).HooksTransferIBCModule.
				ICS4Middleware.Hooks = testutils.TestRecvOverrideHooks{Status: status}
		}, true},
		{"before and after", func(status *testutils.Status) {
			suite.GetNeutronZoneApp(suite.ChainB).HooksTransferIBCModule.
				ICS4Middleware.Hooks = testutils.TestRecvBeforeAfterHooks{Status: status}
		}, true},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			suite.ConfigureTransferChannel()
			receiver = suite.ChainB.SenderAccount.GetAddress().String() // must be explicitly changed in malleate
			status = testutils.Status{}

			amount = sdk.NewInt(100) // must be explicitly changed in malleate
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
				0)
			_, err := suite.ChainA.SendMsgs(transferMsg)
			suite.Require().NoError(err) // message committed

			tc.malleate(&status)

			data := transfertypes.NewFungibleTokenPacketData(trace.GetFullDenomPath(), amount.String(), suite.ChainA.SenderAccount.GetAddress().String(), receiver)
			packet := channeltypes.NewPacket(data.GetBytes(), seq, suite.TransferPath.EndpointA.ChannelConfig.PortID, suite.TransferPath.EndpointA.ChannelID, suite.TransferPath.EndpointB.ChannelConfig.PortID, suite.TransferPath.EndpointB.ChannelID, clienttypes.NewHeight(1, 100), 0)

			ack := suite.GetNeutronZoneApp(suite.ChainB).HooksTransferIBCModule.
				OnRecvPacket(suite.ChainB.GetContext(), packet, suite.ChainA.SenderAccount.GetAddress())

			if tc.expPass {
				suite.Require().True(ack.Success())
			} else {
				suite.Require().False(ack.Success())
			}

			if _, ok := suite.GetNeutronZoneApp(suite.ChainB).HooksTransferIBCModule.
				ICS4Middleware.Hooks.(testutils.TestRecvOverrideHooks); ok {
				suite.Require().True(status.OverrideRan)
				suite.Require().False(status.BeforeRan)
				suite.Require().False(status.AfterRan)
			}

			if _, ok := suite.GetNeutronZoneApp(suite.ChainB).HooksTransferIBCModule.
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
		clienttypes.NewHeight(0, 100),
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

	err := suite.GetNeutronZoneApp(suite.ChainB).HooksICS4Wrapper.SendPacket(
		suite.ChainB.GetContext(), channelCap, packet)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)

	// Update both clients
	err = suite.TransferPath.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.TransferPath.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// recv in chain a
	res, err := suite.TransferPath.EndpointA.RecvPacketWithResult(packet)

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
	codeId := suite.StoreContractCode(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/echo.wasm")
	addr := suite.InstantiateContract(suite.ChainA, sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeId, "{}")

	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")
}

//
//// After successfully executing a wasm call, the contract should have the funds sent via IBC
//func (suite *HooksTestSuite) TestFundsAreTransferredToTheContract() {
//	// Setup contract
//	codeId := suite.chainA.StoreContractCode(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/echo.wasm")
//	addr := suite.chainA.InstantiateContract(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeId, "{}")
//
//	// Check that the contract has no funds
//	localDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
//	balance := suite.chainA.GetNeutronApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
//	suite.Require().Equal(sdk.NewInt(0), balance.Amount)
//
//	// Execute the contract via IBC
//	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"echo": {"msg": "test"} } } }`, addr))
//	ackStr := string(ackBytes)
//	fmt.Println(ackStr)
//	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
//	err := json.Unmarshal(ackBytes, &ack)
//	suite.Require().NoError(err)
//	suite.Require().NotContains(ack, "error")
//	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")
//
//	// Check that the token has now been transferred to the contract
//	balance = suite.chainA.GetNeutronApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
//	suite.Require().Equal(sdk.NewInt(1), balance.Amount)
//}
//
//// If the wasm call wails, the contract acknowledgement should be an error and the funds returned
//func (suite *HooksTestSuite) TestFundsAreReturnedOnFailedContractExec() {
//	// Setup contract
//	codeId := suite.chainA.StoreContractCode(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/echo.wasm")
//	addr := suite.chainA.InstantiateContract(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeId, "{}")
//
//	// Check that the contract has no funds
//	localDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
//	balance := suite.chainA.GetNeutronApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
//	suite.Require().Equal(sdk.NewInt(0), balance.Amount)
//
//	// Execute the contract via IBC with a message that the contract will reject
//	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"not_echo": {"msg": "test"} } } }`, addr))
//	ackStr := string(ackBytes)
//	fmt.Println(ackStr)
//	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
//	err := json.Unmarshal(ackBytes, &ack)
//	suite.Require().NoError(err)
//	suite.Require().Contains(ack, "error")
//
//	// Check that the token has now been transferred to the contract
//	balance = suite.chainA.GetNeutronApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
//	fmt.Println(balance)
//	suite.Require().Equal(sdk.NewInt(0), balance.Amount)
//}
//
//func (suite *HooksTestSuite) TestPacketsThatShouldBeSkipped() {
//	var sequence uint64
//	receiver := suite.chainB.SenderAccount.GetAddress().String()
//
//	testCases := []struct {
//		memo           string
//		expPassthrough bool
//	}{
//		{"", true},
//		{"{01]", true}, // bad json
//		{"{}", true},
//		{`{"something": ""}`, true},
//		{`{"wasm": "test"}`, false},
//		{`{"wasm": []`, true}, // invalid top level JSON
//		{`{"wasm": {}`, true}, // invalid top level JSON
//		{`{"wasm": []}`, false},
//		{`{"wasm": {}}`, false},
//		{`{"wasm": {"contract": "something"}}`, false},
//		{`{"wasm": {"contract": "osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj"}}`, false},
//		{`{"wasm": {"msg": "something"}}`, false},
//		// invalid receiver
//		{`{"wasm": {"contract": "osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj", "msg": {}}}`, false},
//		// msg not an object
//		{fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": 1}}`, receiver), false},
//	}
//
//	for _, tc := range testCases {
//		ackBytes := suite.receivePacketWithSequence(receiver, tc.memo, sequence)
//		ackStr := string(ackBytes)
//		fmt.Println(ackStr)
//		var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
//		err := json.Unmarshal(ackBytes, &ack)
//		suite.Require().NoError(err)
//		if tc.expPassthrough {
//			suite.Require().Equal("AQ==", ack["result"], tc.memo)
//		} else {
//			suite.Require().Contains(ackStr, "error", tc.memo)
//		}
//		sequence += 1
//	}
//}
//
//// After successfully executing a wasm call, the contract should have the funds sent via IBC
//func (suite *HooksTestSuite) TestFundTracking() {
//	// Setup contract
//	codeId := suite.chainA.StoreContractCode(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/counter.wasm")
//	addr := suite.chainA.InstantiateContract(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeId, `{"count": 0}`)
//
//	// Check that the contract has no funds
//	localDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
//	balance := suite.chainA.GetNeutronApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
//	suite.Require().Equal(sdk.NewInt(0), balance.Amount)
//
//	// Execute the contract via IBC
//	suite.receivePacket(
//		addr.String(),
//		fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"increment": {} } } }`, addr))
//
//	senderLocalAcc, err := ibchookskeeper.DeriveIntermediateSender("channel-0", suite.chainB.SenderAccount.GetAddress().String(), "osmo")
//	suite.Require().NoError(err)
//
//	state := suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderLocalAcc)))
//	suite.Require().Equal(`{"count":0}`, state)
//
//	state = suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderLocalAcc)))
//	suite.Require().Equal(`{"total_funds":[{"denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","amount":"1"}]}`, state)
//
//	suite.receivePacketWithSequence(
//		addr.String(),
//		fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"increment": {} } } }`, addr), 1)
//
//	state = suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderLocalAcc)))
//	suite.Require().Equal(`{"count":1}`, state)
//
//	state = suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderLocalAcc)))
//	suite.Require().Equal(`{"total_funds":[{"denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","amount":"2"}]}`, state)
//
//	// Check that the token has now been transferred to the contract
//	balance = suite.chainA.GetNeutronApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
//	suite.Require().Equal(sdk.NewInt(2), balance.Amount)
//}

// custom MsgTransfer constructor that supports Memo
func NewMsgTransfer(
	token sdk.Coin, sender, receiver string, memo string,
) *transfertypes.MsgTransfer {
	return &transfertypes.MsgTransfer{
		SourcePort:       "transfer",
		SourceChannel:    "channel-0",
		Token:            token,
		Sender:           sender,
		Receiver:         receiver,
		TimeoutHeight:    clienttypes.NewHeight(0, 100),
		TimeoutTimestamp: 0,
		Memo:             memo,
	}
}

type Direction int64

const (
	AtoB Direction = iota
	BtoA
)

func (suite *HooksTestSuite) GetEndpoints(direction Direction) (sender *ibctesting.Endpoint, receiver *ibctesting.Endpoint) {
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

func (suite *HooksTestSuite) RelayPacket(packet channeltypes.Packet, direction Direction) (*sdk.Result, []byte) {
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

//func (suite *HooksTestSuite) FullSend(msg sdk.Msg, direction Direction) (*sdk.Result, *sdk.Result, string, error) {
//	var sender *TestChain
//	switch direction {
//	case AtoB:
//		sender = suite.ChainA
//	case BtoA:
//		sender = suite.ChainB
//	}
//	sendResult, err := sender.SendMsgsNoCheck(msg)
//	suite.Require().NoError(err)
//
//	packet, err := ParsePacketFromEvents(sendResult.GetEvents())
//	suite.Require().NoError(err)
//
//	receiveResult, ack := suite.RelayPacket(packet, direction)
//
//	return sendResult, receiveResult, string(ack), err
//}

//func (suite *HooksTestSuite) TestAcks() {
//	suite.ConfigureTransferChannel()
//
//	// Setup contract
//	codeId := suite.chainA.StoreContractCode(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/counter.wasm")
//	addr := suite.chainA.InstantiateContract(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeId, `{"count": 0}`)
//
//	// Generate swap instructions for the contract
//	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)
//	// Send IBC transfer with the memo with crosschain-swap instructions
//	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), addr.String(), callbackMemo)
//	suite.FullSend(transferMsg, AtoB)
//
//	// The test contract will increment the counter for itself every time it receives an ack
//	state := suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
//	suite.Require().Equal(`{"count":1}`, state)
//
//	suite.FullSend(transferMsg, AtoB)
//	state = suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
//	suite.Require().Equal(`{"count":2}`, state)
//
//}

//func (suite *HooksTestSuite) TestTimeouts() {
//	// Setup contract
//	codeId := suite.chainA.StoreContractCode(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), "./bytecode/counter.wasm")
//	addr := suite.chainA.InstantiateContract(suite.chainA.GetContext(), sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress), codeId, `{"count": 0}`)
//
//	// Generate swap instructions for the contract
//	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)
//	// Send IBC transfer with the memo with crosschain-swap instructions
//	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), addr.String(), callbackMemo)
//	transferMsg.TimeoutTimestamp = uint64(suite.coordinator.CurrentTime.Add(time.Minute).UnixNano())
//	sendResult, err := suite.chainA.SendMsgsNoCheck(transferMsg)
//	suite.Require().NoError(err)
//
//	packet, err := ParsePacketFromEvents(sendResult.GetEvents())
//	suite.Require().NoError(err)
//
//	// Move chainB forward one block
//	suite.chainB.NextBlock()
//	// One month later
//	suite.coordinator.IncrementTimeBy(time.Hour)
//	err = suite.path.EndpointA.UpdateClient()
//	suite.Require().NoError(err)
//
//	err = suite.path.EndpointA.TimeoutPacket(packet)
//	suite.Require().NoError(err)
//
//	// The test contract will increment the counter for itself by 10 when a packet times out
//	state := suite.chainA.QueryContract(
//		suite.chainA.GetContext(),
//		addr,
//		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
//	suite.Require().Equal(`{"count":10}`, state)
//
//}

//func (suite *HooksTestSuite) TestSendWithoutMemo() {
//	// Sending a packet without memo to ensure that the ibc_callback middleware doesn't interfere with a regular send
//	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), "")
//	_, _, ack, err := suite.FullSend(transferMsg, AtoB)
//	suite.Require().NoError(err)
//	suite.Require().Contains(ack, "result")
//}

type Chain int64

const (
	ChainA Chain = iota
	ChainB
)

//func (suite *HooksTestSuite) SetupIBCRouteOnChainB(swaprouterAddr, owner sdk.AccAddress) {
//	chain := suite.GetChain(ChainB)
//	denomTrace1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token1"))
//	token1IBC := denomTrace1.IBCDenom()
//
//	msg := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"token0","pool_route":[{"pool_id":"3","token_out_denom":"stake"},{"pool_id":"1","token_out_denom":"token0"}]}}`,
//		token1IBC)
//	osmosisApp := chain.GetNeutronApp()
//	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
//	_, err := contractKeeper.Execute(chain.GetContext(), swaprouterAddr, owner, []byte(msg), sdk.NewCoins())
//	suite.Require().NoError(err)
//
//	// Move forward one block
//	chain.NextBlock()
//	chain.Coordinator.IncrementTime()
//
//	// Update both clients
//	err = suite.path.EndpointA.UpdateClient()
//	suite.Require().NoError(err)
//	err = suite.path.EndpointB.UpdateClient()
//	suite.Require().NoError(err)
//
//}

//func (suite *HooksTestSuite) GetChain(name Chain) *TestChain {
//	if name == ChainA {
//		return suite.chainA
//	} else {
//		return suite.chainB
//	}
//}
//
//type TestChain struct {
//	*ibctesting.TestChain
//}

// SendMsgsNoCheck overrides ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func SendMsgsNoCheck(chain *ibctesting.TestChain, msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	_, r, err := SignAndDeliver(
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliver(
	txCfg client.TxConfig, app *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, _ := helpers.GenTx(
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		helpers.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)

	// Simulate a sending a transaction and committing a block
	gInfo, res, err := app.Deliver(txCfg.TxEncoder(), tx)

	return gInfo, res, err
}

func (suite *HooksTestSuite) StoreContractCode(chain *ibctesting.TestChain, addr sdk.AccAddress, path string) uint64 {
	wasmCode, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	codeID, _, err := wasmkeeper.NewDefaultPermissionKeeper(suite.GetNeutronZoneApp(chain).WasmKeeper).Create(chain.GetContext(), addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody, Address: ""})
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

// ParsePacketFromEvents parses events emitted from a MsgRecvPacket and returns the
// acknowledgement.
func ParsePacketFromEvents(events sdk.Events) (channeltypes.Packet, error) {
	for _, ev := range events {
		if ev.Type == channeltypes.EventTypeSendPacket {
			packet := channeltypes.Packet{}
			for _, attr := range ev.Attributes {
				switch string(attr.Key) {
				case channeltypes.AttributeKeyData:
					packet.Data = attr.Value

				case channeltypes.AttributeKeySequence:
					seq, err := strconv.ParseUint(string(attr.Value), 10, 64)
					if err != nil {
						return channeltypes.Packet{}, err
					}

					packet.Sequence = seq

				case channeltypes.AttributeKeySrcPort:
					packet.SourcePort = string(attr.Value)

				case channeltypes.AttributeKeySrcChannel:
					packet.SourceChannel = string(attr.Value)

				case channeltypes.AttributeKeyDstPort:
					packet.DestinationPort = string(attr.Value)

				case channeltypes.AttributeKeyDstChannel:
					packet.DestinationChannel = string(attr.Value)

				case channeltypes.AttributeKeyTimeoutHeight:
					height, err := clienttypes.ParseHeight(string(attr.Value))
					if err != nil {
						return channeltypes.Packet{}, err
					}

					packet.TimeoutHeight = height

				case channeltypes.AttributeKeyTimeoutTimestamp:
					timestamp, err := strconv.ParseUint(string(attr.Value), 10, 64)
					if err != nil {
						return channeltypes.Packet{}, err
					}

					packet.TimeoutTimestamp = timestamp

				default:
					continue
				}
			}

			return packet, nil
		}
	}
	return channeltypes.Packet{}, fmt.Errorf("acknowledgement event attribute not found")
}
