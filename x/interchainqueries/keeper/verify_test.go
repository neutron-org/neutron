package keeper_test

import (
	"fmt"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmprotoversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
	tmversion "github.com/tendermint/tendermint/version"
	"time"
)

// CreateTMClientHeader creates a TM header to update the TM client. Args are passed in to allow
// caller flexibility to use params that differ from the chain.
func CreateTMClientHeader(chain *ibctesting.TestChain, chainID string, blockHeight int64, trustedHeight ibcclienttypes.Height, timestamp time.Time, tmValSet, tmTrustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator, previousHeader *tmtypes.Header) *ibctmtypes.Header {
	var (
		valSet      *tmproto.ValidatorSet
		trustedVals *tmproto.ValidatorSet
	)
	require.NotNil(chain.T, tmValSet)

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

	commit, err := tmtypes.MakeCommit(blockID, blockHeight, 1, voteSet, signers, timestamp)
	require.NoError(chain.T, err)

	signedHeader := &tmproto.SignedHeader{
		Header: tmHeader.ToProto(),
		Commit: commit.ToProto(),
	}

	if tmValSet != nil {
		valSet, err = tmValSet.ToProto()
		require.NoError(chain.T, err)
	}

	if tmTrustedVals != nil {
		trustedVals, err = tmTrustedVals.ToProto()
		require.NoError(chain.T, err)
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
	// set the last header to the current header
	// use nil trusted fields
	ph, err := tmtypes.HeaderFromProto(chain.LastHeader.Header)
	require.NoError(chain.T, err)

	chain.LastHeader = CreateTMClientHeader(chain, chain.ChainID, chain.CurrentHeader.Height, ibcclienttypes.Height{}, chain.CurrentHeader.Time, chain.Vals, nil, chain.Signers, &ph)

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

	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
}

// CommitBlock commits a block on the provided indexes and then increments the global time.
//
// CONTRACT: the passed in list of indexes must not contain duplicates
func CommitBlock(coord *ibctesting.Coordinator, chains ...*ibctesting.TestChain) {
	for _, chain := range chains {
		chain.App.Commit()
		NextBlock(chain)
	}
	coord.IncrementTime()
}

// UpdateClient updates the IBC client associated with the endpoint.
func UpdateClient(endpoint *ibctesting.Endpoint) (err error) {
	var header exported.Header

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
	require.NoError(endpoint.Chain.T, err)

	_, err = endpoint.Chain.SendMsgs(msg)

	return err
}

func (suite *KeeperTestSuite) TestUnpackAndVerifyHeaders() {
	var (
		header     ibctmtypes.Header
		nextHeader ibctmtypes.Header
		clientID   string
	)

	tests := []struct {
		name      string
		malleate  func()
		expectErr bool
	}{
		{
			"valid headers",
			func() {
				suite.Require().NoError(UpdateClient(suite.path.EndpointA))

				clientID = suite.path.EndpointA.ClientID

				h, err := suite.path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.path.EndpointA.Counterparty.Chain, suite.path.EndpointB.ClientID)
				suite.Require().NoError(err)
				header = *h

				CommitBlock(suite.coordinator, suite.chainB)
				nh, err := suite.path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.path.EndpointA.Counterparty.Chain, suite.path.EndpointB.ClientID)
				suite.Require().NoError(err)
				nextHeader = *nh
			},
			false,
		},
		{
			"headers are not sequential",
			func() {
				suite.Require().NoError(UpdateClient(suite.path.EndpointA))

				clientID = suite.path.EndpointA.ClientID

				h, err := suite.path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.path.EndpointA.Counterparty.Chain, suite.path.EndpointB.ClientID)
				suite.Require().NoError(err)
				header = *h

				// skip one block to set nextHeader's height + 2
				CommitBlock(suite.coordinator, suite.chainB)
				CommitBlock(suite.coordinator, suite.chainB)

				nh, err := suite.path.EndpointA.Chain.ConstructUpdateTMClientHeader(suite.path.EndpointA.Counterparty.Chain, suite.path.EndpointB.ClientID)
				suite.Require().NoError(err)
				nextHeader = *nh
			},
			true,
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i, len(tests)), func() {
			suite.SetupTest()

			tt.malleate()

			err := suite.GetGaiaWasmZoneApp(suite.chainA).InterchainQueriesKeeper.VerifyHeaders(suite.chainA.GetContext(), clientID, &header, &nextHeader)

			if tt.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
