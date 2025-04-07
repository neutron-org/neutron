package keeper_test

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/golang/mock/gomock"

	icqtestkeeper "github.com/neutron-org/neutron/v6/testutil/interchainqueries/keeper"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/interchainqueries/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmprotoversion "github.com/cometbft/cometbft/proto/tendermint/version"
	tmtypes "github.com/cometbft/cometbft/types"
	tmversion "github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"

	"github.com/neutron-org/neutron/v6/testutil"
	iqkeeper "github.com/neutron-org/neutron/v6/x/interchainqueries/keeper"
	iqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

// CreateTMClientHeader creates a TM header to update the TM client. Args are passed in to allow
// caller flexibility to use params that differ from the chain.
func CreateTMClientHeader(chain *ibctesting.TestChain, chainID string, blockHeight int64, trustedHeight ibcclienttypes.Height, timestamp time.Time, tmValSet, tmTrustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator, previousHeader *tmtypes.Header) *ibctmtypes.Header {
	var (
		valSet      *tmproto.ValidatorSet
		trustedVals *tmproto.ValidatorSet
	)
	require.NotNil(chain.TB, tmValSet)

	vsetHash := tmValSet.Hash()

	tmHeader := tmtypes.Header{
		Version:            tmprotoversion.Consensus{Block: tmversion.BlockProtocol, App: 2},
		ChainID:            chainID,
		Height:             blockHeight,
		Time:               timestamp,
		LastBlockID:        ibctesting.MakeBlockID(previousHeader.Hash(), 10_000, make([]byte, tmhash.Size)),
		LastCommitHash:     chain.App.LastCommitID().Hash,
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            chain.CurrentHeader.AppHash,
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    tmValSet.Proposer.Address, //nolint:staticcheck
	}

	hhash := tmHeader.Hash()
	blockID := ibctesting.MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet(chainID, blockHeight, 1, tmproto.PrecommitType, tmValSet)

	commit, err := tmtypes.MakeExtCommit(blockID, blockHeight, 1, voteSet, signers, timestamp, false)
	require.NoError(chain.TB, err)

	signedHeader := &tmproto.SignedHeader{
		Header: tmHeader.ToProto(),
		Commit: commit.ToCommit().ToProto(),
	}

	if tmValSet != nil { //nolint:staticcheck // this checks if a pointer is nil, suggesting that it can be nil but we have this test all over the place
		valSet, err = tmValSet.ToProto()
		require.NoError(chain.TB, err)
	}

	if tmTrustedVals != nil {
		trustedVals, err = tmTrustedVals.ToProto()
		require.NoError(chain.TB, err)
	}

	// The trusted fields may be nil. They may be filled before relaying messages to a client.
	// The relayer is responsible for querying client and injecting appropriate trusted fields.
	return &ibctmtypes.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      valSet,
		TrustedHeight:     trustedHeight,
		TrustedValidators: trustedVals,
	}
}

func NextBlock(chain *ibctesting.TestChain) {
	_, err := chain.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             chain.CurrentHeader.Height,
		Time:               chain.CurrentHeader.GetTime(),
		NextValidatorsHash: chain.NextVals.Hash(),
	})
	require.NoError(chain.TB, err)

	_, err = chain.App.Commit()
	require.NoError(chain.TB, err)

	// set the last header to the current header
	// use nil trusted fields
	ph, err := tmtypes.HeaderFromProto(chain.LastHeader.Header)
	require.NoError(chain.TB, err)

	var signers []tmtypes.PrivValidator
	for _, val := range chain.Vals.Validators {
		signers = append(signers, chain.Signers[val.PubKey.Address().String()])
	}
	chain.LastHeader = CreateTMClientHeader(chain, chain.ChainID, chain.CurrentHeader.Height, ibcclienttypes.Height{}, chain.CurrentHeader.Time, chain.Vals, nil, signers, &ph)

	// increment the current header
	chain.CurrentHeader = tmproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.Vals.Hash(),
	}
}

// CommitBlock commits a block on the provided indexes and then increments the global time.
//
// CONTRACT: the passed in list of indexes must not contain duplicates
func CommitBlock(coord *ibctesting.Coordinator, chains ...*ibctesting.TestChain) {
	for _, chain := range chains {
		NextBlock(chain)
	}
	coord.IncrementTime()
}

// UpdateClient updates the IBC client associated with the endpoint.
func UpdateClient(endpoint *ibctesting.Endpoint) (err error) {
	var header exported.ClientMessage

	// ensure counterparty has committed state
	CommitBlock(endpoint.Chain.Coordinator, endpoint.Counterparty.Chain)

	switch endpoint.ClientConfig.GetClientType() {
	case exported.Tendermint:
		header, err = endpoint.Chain.ConstructUpdateTMClientHeader(endpoint.Counterparty.Chain, endpoint.ClientID)

	default:
		err = fmt.Errorf("client type %s is not supported", endpoint.ClientConfig.GetClientType())
	}

	if err != nil {
		return err
	}

	msg, err := ibcclienttypes.NewMsgUpdateClient(
		endpoint.ClientID, header,
		endpoint.Chain.SenderAccount.GetAddress().String(),
	)
	require.NoError(endpoint.Chain.TB, err)

	_, err = endpoint.Chain.SendMsgs(msg)

	return err
}

func (suite *KeeperTestSuite) TestUnpackAndVerifyHeaders() {
	tests := []struct {
		name             string
		run              func() error
		expectedErrorMsg string
	}{
		{
			"valid headers",
			func() error {
				suite.Require().NoError(suite.Path.EndpointA.UpdateClient())

				clientID := suite.Path.EndpointA.ClientID
				CommitBlock(suite.Coordinator, suite.ChainB)
				header, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID)
				suite.Require().NoError(err)

				CommitBlock(suite.Coordinator, suite.ChainB)
				nextHeader, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID)
				suite.Require().NoError(err)

				return iqkeeper.Verifier{}.VerifyHeaders(suite.ChainA.GetContext(), suite.GetNeutronZoneApp(suite.ChainA).IBCKeeper.ClientKeeper, clientID, header, nextHeader)
			},
			"",
		},
		{
			"headers are not sequential",
			func() error {
				suite.Require().NoError(UpdateClient(suite.Path.EndpointA))

				clientID := suite.Path.EndpointA.ClientID
				CommitBlock(suite.Coordinator, suite.ChainB)

				header, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID)
				suite.Require().NoError(err)

				// skip one block to set nextHeader's height + 2
				CommitBlock(suite.Coordinator, suite.ChainB)
				CommitBlock(suite.Coordinator, suite.ChainB)

				nextHeader, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID)
				suite.Require().NoError(err)

				return iqkeeper.Verifier{}.VerifyHeaders(suite.ChainA.GetContext(), suite.GetNeutronZoneApp(suite.ChainA).IBCKeeper.ClientKeeper, clientID, header, nextHeader)
			},
			"block.NextBlockHeader is not next for the block.Header",
		},
		{
			"header has some malicious field",
			func() error {
				suite.Require().NoError(UpdateClient(suite.Path.EndpointA))

				clientID := suite.Path.EndpointA.ClientID
				CommitBlock(suite.Coordinator, suite.ChainB)
				CommitBlock(suite.Coordinator, suite.ChainB)

				header, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID)
				suite.Require().NoError(err)

				CommitBlock(suite.Coordinator, suite.ChainB)

				header.SignedHeader.Header.LastResultsHash = []byte("malicious hash with length 32!!!")

				nextHeader, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID)
				suite.Require().NoError(err)

				return iqkeeper.Verifier{}.VerifyHeaders(suite.ChainA.GetContext(), suite.GetNeutronZoneApp(suite.ChainA).IBCKeeper.ClientKeeper, clientID, header, nextHeader)
			},
			"invalid header: untrustedHeader.ValidateBasic failed: commit signs block",
		},
		{
			"headers from the past (when client on chain A has the most recent consensus state and relayer try to submit old headers from chain B)",
			func() error {
				suite.Require().NoError(UpdateClient(suite.Path.EndpointA))

				clientID := suite.Path.EndpointA.ClientID
				CommitBlock(suite.Coordinator, suite.ChainB)

				oldHeader := *suite.ChainB.LastHeader
				CommitBlock(suite.Coordinator, suite.ChainB)
				oldNextHeader := *suite.ChainB.LastHeader

				for i := 0; i < 30; i++ {
					suite.Require().NoError(UpdateClient(suite.Path.EndpointA))
				}
				headerWithTrustedHeight, err := suite.Path.EndpointA.Chain.ConstructUpdateTMClientHeaderWithTrustedHeight(suite.Path.EndpointA.Counterparty.Chain, suite.Path.EndpointB.ClientID, ibcclienttypes.Height{
					RevisionNumber: 0,
					RevisionHeight: 15,
				})
				suite.Require().NoError(err)

				oldHeader.TrustedHeight = headerWithTrustedHeight.TrustedHeight
				oldHeader.TrustedValidators = headerWithTrustedHeight.TrustedValidators

				oldNextHeader.TrustedHeight = headerWithTrustedHeight.TrustedHeight
				oldNextHeader.TrustedValidators = headerWithTrustedHeight.TrustedValidators

				return iqkeeper.Verifier{}.VerifyHeaders(suite.ChainA.GetContext(), suite.GetNeutronZoneApp(suite.ChainA).IBCKeeper.ClientKeeper, clientID, &oldHeader, &oldNextHeader)
			},
			"",
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			var (
				ctx           = suite.ChainA.GetContext()
				contractOwner = keeper.RandomAccountAddress(suite.T()) // We don't care what this address is
			)

			// Store code and instantiate reflect contract.
			codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
			contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
			suite.Require().NotEmpty(contractAddress)

			err := testutil.SetupICAPath(suite.Path, contractAddress.String())
			suite.Require().NoError(err)

			err = tt.run()
			if tt.expectedErrorMsg != "" {
				suite.Require().ErrorContains(err, tt.expectedErrorMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func TestSudoHasAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hv := mock_types.NewMockHeaderVerifier(ctrl)
	tv := mock_types.NewMockTransactionVerifier(ctrl)
	cm := mock_types.NewMockContractManagerKeeper(ctrl)
	ibck := ibckeeper.Keeper{ClientKeeper: clientkeeper.Keeper{}}

	k, ctx := icqtestkeeper.InterchainQueriesKeeper(t, &ibck, cm, hv, tv)
	address := types.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	header := ibctmtypes.Header{
		SignedHeader: &tmproto.SignedHeader{
			Header: &tmproto.Header{Height: 1001},
		},
		TrustedHeight: ibcclienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 1001,
		},
	}
	nextHeader := ibctmtypes.Header{
		TrustedHeight: ibcclienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 1002,
		},
	}
	packedHeader, err := codectypes.NewAnyWithValue(&header)
	require.NoError(t, err)
	packedNextHeader, err := codectypes.NewAnyWithValue(&nextHeader)
	require.NoError(t, err)
	tx := iqtypes.TxValue{
		Response:       nil,
		DeliveryProof:  nil,
		InclusionProof: nil,
		Data:           []byte("txbody"),
	}
	block := iqtypes.Block{
		NextBlockHeader: packedNextHeader,
		Header:          packedHeader,
		Tx:              &tx,
	}

	hv.EXPECT().UnpackHeader(packedHeader).Return(nil, fmt.Errorf("failed to unpack packedHeader"))
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.ErrorContains(t, err, "failed to unpack block header")

	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(nil, fmt.Errorf("failed to unpack packedHeader"))
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.ErrorContains(t, err, "failed to unpack next block header")

	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(exported.ClientMessage(&nextHeader), nil)
	hv.EXPECT().VerifyHeaders(ctx, clientkeeper.Keeper{}, "tendermint-07", exported.ClientMessage(&header), exported.ClientMessage(&nextHeader)).Return(fmt.Errorf("failed to verify headers"))
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.ErrorContains(t, err, "failed to verify headers")

	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(exported.ClientMessage(&nextHeader), nil)
	hv.EXPECT().VerifyHeaders(ctx, clientkeeper.Keeper{}, "tendermint-07", exported.ClientMessage(&header), exported.ClientMessage(&nextHeader)).Return(nil)
	tv.EXPECT().VerifyTransaction(&header, &nextHeader, &tx).Return(fmt.Errorf("failed to verify transaction"))
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.ErrorContains(t, err, "failed to verifyTransaction")

	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(exported.ClientMessage(&nextHeader), nil)
	hv.EXPECT().VerifyHeaders(ctx, clientkeeper.Keeper{}, "tendermint-07", exported.ClientMessage(&header), exported.ClientMessage(&nextHeader)).Return(nil)
	tv.EXPECT().VerifyTransaction(&header, &nextHeader, &tx).Return(nil)
	cm.EXPECT().SudoTxQueryResult(ctx, address, uint64(1), ibcclienttypes.NewHeight(1, uint64(header.Header.Height)), tx.GetData()).Return(nil, fmt.Errorf("contract error")) //nolint:gosec
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.ErrorContains(t, err, "rejected transaction query result")

	// all error flows passed, time to success
	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(exported.ClientMessage(&nextHeader), nil)
	hv.EXPECT().VerifyHeaders(ctx, clientkeeper.Keeper{}, "tendermint-07", exported.ClientMessage(&header), exported.ClientMessage(&nextHeader)).Return(nil)
	tv.EXPECT().VerifyTransaction(&header, &nextHeader, &tx).Return(nil)
	cm.EXPECT().SudoTxQueryResult(ctx, address, uint64(1), ibcclienttypes.NewHeight(1, uint64(header.Header.Height)), tx.GetData()).Return(nil, nil) //nolint:gosec
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.NoError(t, err)

	// no functions calls after VerifyHeaders means we try to process tx second time
	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(exported.ClientMessage(&nextHeader), nil)
	hv.EXPECT().VerifyHeaders(ctx, clientkeeper.Keeper{}, "tendermint-07", exported.ClientMessage(&header), exported.ClientMessage(&nextHeader)).Return(nil)
	err = k.ProcessBlock(ctx, address, 1, "tendermint-07", &block)
	require.NoError(t, err)

	// same tx + another queryID
	hv.EXPECT().UnpackHeader(packedHeader).Return(exported.ClientMessage(&header), nil)
	hv.EXPECT().UnpackHeader(packedNextHeader).Return(exported.ClientMessage(&nextHeader), nil)
	hv.EXPECT().VerifyHeaders(ctx, clientkeeper.Keeper{}, "tendermint-07", exported.ClientMessage(&header), exported.ClientMessage(&nextHeader)).Return(nil)
	tv.EXPECT().VerifyTransaction(&header, &nextHeader, &tx).Return(nil)
	cm.EXPECT().SudoTxQueryResult(ctx, address, uint64(2), ibcclienttypes.NewHeight(1, uint64(header.Header.Height)), tx.GetData()).Return(nil, nil) //nolint:gosec
	err = k.ProcessBlock(ctx, address, 2, "tendermint-07", &block)
	require.NoError(t, err)
}
