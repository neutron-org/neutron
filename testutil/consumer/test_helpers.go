package consumer

import (
	"encoding/json"
	"time"

	"cosmossdk.io/errors"

	types1 "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ccvconsumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"
	ccvprovidertypes "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"
	"github.com/cosmos/interchain-security/v4/x/ccv/types"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/neutron-org/neutron/v3/app"
)

// This function creates consumer module genesis state that is used as starting point for modifications
// that allow neutron chain to be started locally without having to start the provider chain and the relayer.
// It is also used in tests that are starting the chain node.
func CreateMinimalConsumerTestGenesis() *ccvconsumertypes.GenesisState {
	genesisState := ccvconsumertypes.DefaultGenesisState()
	genesisState.Params.Enabled = true
	genesisState.NewChain = true
	genesisState.Provider.ClientState = ccvprovidertypes.DefaultParams().TemplateClient
	genesisState.Provider.ClientState.ChainId = app.Name
	genesisState.Provider.ClientState.LatestHeight = ibcclienttypes.Height{RevisionNumber: 0, RevisionHeight: 1}
	genesisState.Provider.ClientState.TrustingPeriod, _ = types.CalculateTrustPeriod(genesisState.Params.UnbondingPeriod, ccvprovidertypes.DefaultTrustingPeriodFraction)

	genesisState.Provider.ClientState.UnbondingPeriod = genesisState.Params.UnbondingPeriod
	genesisState.Provider.ClientState.MaxClockDrift = ccvprovidertypes.DefaultMaxClockDrift
	genesisState.Provider.ConsensusState = &ibctmtypes.ConsensusState{
		Timestamp: time.Now().UTC(),
		Root:      ibccommitmenttypes.MerkleRoot{Hash: []byte("dummy")},
	}

	return genesisState
}

func ModifyConsumerGenesis(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errors.Wrap(err, "failed to read genesis from the file")
	}

	tmProtoPublicKey, err := cryptocodec.ToTmProtoPublicKey(val.PubKey)
	if err != nil {
		return errors.Wrap(err, "invalid public key")
	}

	initialValset := []types1.ValidatorUpdate{{PubKey: tmProtoPublicKey, Power: 100}}
	vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValset)
	if err != nil {
		return errors.Wrap(err, "could not convert val updates to validator set")
	}

	consumerGenesisState := CreateMinimalConsumerTestGenesis()
	consumerGenesisState.Provider.InitialValSet = initialValset
	consumerGenesisState.Provider.ConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).Hash()

	if err := consumerGenesisState.Validate(); err != nil {
		return errors.Wrap(err, "invalid consumer genesis")
	}

	consumerGenStateBz, err := val.ClientCtx.Codec.MarshalJSON(consumerGenesisState)
	if err != nil {
		return errors.Wrap(err, "failed to marshal consumer genesis state into JSON")
	}

	appState[ccvconsumertypes.ModuleName] = consumerGenStateBz
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return errors.Wrap(err, "failed to marshal application genesis state into JSON")
	}

	genDoc.AppState = appStateJSON
	err = genutil.ExportGenesisFile(genDoc, genFile)
	if err != nil {
		return errors.Wrap(err, "failed to export genesis state")
	}

	return nil
}

func ModifyGenesisBlockGas(val network.Validator) error {
	genFile := val.Ctx.Config.GenesisFile()
	_, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return errors.Wrap(err, "failed to read genesis from the file")
	}

	genDoc.ConsensusParams.Block.MaxGas = 35_000_000

	err = genutil.ExportGenesisFile(genDoc, genFile)
	if err != nil {
		return errors.Wrap(err, "failed to export genesis state")
	}

	return nil
}
