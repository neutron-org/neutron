package stateverifier_test

import (
	"testing"
	"time"

	ibccommitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	tendermint "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v8/testutil/apptesting"
	stateverifier "github.com/neutron-org/neutron/v8/x/state-verifier"
	"github.com/neutron-org/neutron/v8/x/state-verifier/types"
)

type GenesisTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *GenesisTestSuite) TestInitExportGenesis() {
	states := []*types.ConsensusState{
		{
			Height: 1,
			Cs: &tendermint.ConsensusState{
				Timestamp:          time.Now().UTC(),
				Root:               ibccommitmenttypes.MerkleRoot{Hash: []byte("MerkleRoot")},
				NextValidatorsHash: []byte("qqqqqqqqq"),
			},
		},
		{
			Height: 2,
			Cs: &tendermint.ConsensusState{
				Timestamp:          time.Now().UTC(),
				Root:               ibccommitmenttypes.MerkleRoot{Hash: []byte("oafpaosfsdf")},
				NextValidatorsHash: []byte("sdfsdfsdf"),
			},
		},
		{
			Height: 3,
			Cs: &tendermint.ConsensusState{
				Timestamp:          time.Now().UTC(),
				Root:               ibccommitmenttypes.MerkleRoot{Hash: []byte("okjdfhjsdfsdf")},
				NextValidatorsHash: []byte("irhweiriweyrwe"),
			},
		},
	}
	suite.SetupTest()
	k := suite.App.StateVerifierKeeper

	initialGenesis := types.GenesisState{
		States: states,
	}

	stateverifier.InitGenesis(suite.Ctx, k, initialGenesis)

	exportedGenesis := stateverifier.ExportGenesis(suite.Ctx, k)

	suite.Require().EqualValues(initialGenesis, *exportedGenesis)
}
