package main

import (
	"cosmossdk.io/errors"
	"encoding/json"
	"fmt"
	"github.com/neutron-org/neutron/v3/testutil/consumer"
	"os"
	"strings"

	types1 "github.com/cometbft/cometbft/abci/types"
	pvm "github.com/cometbft/cometbft/privval"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"
	"github.com/spf13/cobra"
)

func AddConsumerSectionCmd(defaultNodeHome string) *cobra.Command {
	genesisMutator := NewDefaultGenesisIO()

	txCmd := &cobra.Command{
		Use:                        "add-consumer-section",
		Short:                      "ONLY FOR TESTING PURPOSES! Modifies genesis so that chain can be started locally with one node.",
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			return genesisMutator.AlterConsumerModuleState(cmd, func(state *GenesisData, _ map[string]json.RawMessage) error {
				genesisState := consumer.CreateMinimalConsumerTestGenesis()
				clientCtx := client.GetClientContextFromCmd(cmd)
				serverCtx := server.GetServerContextFromCmd(cmd)
				config := serverCtx.Config
				config.SetRoot(clientCtx.HomeDir)

				valDirs := []string{"val_a", "val_b", "val_c", "val_d"}

				runnerVal, err := cmd.Flags().GetString("validator")
				if err != nil {
					return err
				}

				var initialValset []types1.ValidatorUpdate
				//var peerIds []string
				for _, valDir := range valDirs {
					privValidator := pvm.LoadFilePVEmptyState("/opt/neutron/vals/"+valDir, "")
					pk, err := privValidator.GetPubKey()
					if err != nil {
						return err
					}
					sdkPublicKey, err := cryptocodec.FromTmPubKeyInterface(pk)
					if err != nil {
						return err
					}
					tmProtoPublicKey, err := cryptocodec.ToTmProtoPublicKey(sdkPublicKey)
					if err != nil {
						return err
					}
					initialValset = append(initialValset, types1.ValidatorUpdate{PubKey: tmProtoPublicKey, Power: 25})

					// ====== get peer ids ======

					//if runnerVal != valDir {
					//	nodeKey, err := p2p.LoadNodeKey(valDir + "/config/node_key.json")
					//	if err != nil {
					//		return err
					//	}
					//	peerIds = append(peerIds, string(nodeKey.ID()))
					//}
				}

				vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValset)
				if err != nil {
					return errors.Wrap(err, "could not convert val updates to validator set")
				}

				err2 := writePeersIntoCargo(err, runnerVal)
				if err2 != nil {
					return err2
				}

				genesisState.Provider.InitialValSet = initialValset
				genesisState.Provider.ConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).Hash()

				state.ConsumerModuleState = genesisState

				return nil
			})
		},
	}

	txCmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(txCmd)

	txCmd.Flags().String("validator", defaultNodeHome, "Validator folder")
	flags.AddQueryFlagsToCmd(txCmd)

	return txCmd
}

func writePeersIntoCargo(err error, runnerVal string) error {
	// TODO: write peer ids into config.toml before start
	var data []byte
	data, err = os.ReadFile("/opt/neutron/peers.json")
	var peers [][]string
	if err := json.Unmarshal(data, &peers); err != nil {
		return err
	}
	var res []string
	for _, peer := range peers {
		if peer[0] != runnerVal {
			res = append(res, peer[1])
		}
	}
	peersStr := strings.Join(res, ",")

	baseConfigBytes, err := os.ReadFile("config/config.toml")
	if err != nil {
		return err
	}
	baseConfig := strings.Replace(string(baseConfigBytes), "seeds = \"\"", "seeds = \""+peersStr+"\"", -1)
	baseConfig = strings.Replace(baseConfig, "persistent_peers = \"\"", "persistent_peers = \""+peersStr+"\"", -1)
	err = os.WriteFile("/opt/neutron/data/config/config.toml", []byte(baseConfig), 0644)
	if err != nil {
		return err
	}

	return nil
}

type DefaultGenesisIO struct {
	DefaultGenesisReader
}

func NewDefaultGenesisIO() *DefaultGenesisIO {
	return &DefaultGenesisIO{DefaultGenesisReader: DefaultGenesisReader{}}
}

func (x DefaultGenesisIO) AlterConsumerModuleState(cmd *cobra.Command, callback func(state *GenesisData, appState map[string]json.RawMessage) error) error {
	g, err := x.ReadGenesis(cmd)
	if err != nil {
		return err
	}
	if err := callback(g, g.AppState); err != nil {
		return err
	}
	if err := g.ConsumerModuleState.Validate(); err != nil {
		return err
	}
	clientCtx := client.GetClientContextFromCmd(cmd)
	consumerGenStateBz, err := clientCtx.Codec.MarshalJSON(g.ConsumerModuleState)
	if err != nil {
		return errors.Wrap(err, "marshal consumer genesis state")
	}

	g.AppState[ccvconsumertypes.ModuleName] = consumerGenStateBz
	appStateJSON, err := json.Marshal(g.AppState)
	if err != nil {
		return errors.Wrap(err, "marshal application genesis state")
	}

	g.GenDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(g.GenDoc, g.GenesisFile)
}

type DefaultGenesisReader struct{}

func (d DefaultGenesisReader) ReadGenesis(cmd *cobra.Command) (*GenesisData, error) {
	clientCtx := client.GetClientContextFromCmd(cmd)
	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config
	config.SetRoot(clientCtx.HomeDir)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	return NewGenesisData(
		genFile,
		genDoc,
		appState,
		nil,
	), nil
}

type GenesisData struct {
	GenesisFile         string
	GenDoc              *tmtypes.GenesisDoc
	AppState            map[string]json.RawMessage
	ConsumerModuleState *ccvconsumertypes.GenesisState
}

func NewGenesisData(genesisFile string, genDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage, consumerModuleState *ccvconsumertypes.GenesisState) *GenesisData {
	return &GenesisData{GenesisFile: genesisFile, GenDoc: genDoc, AppState: appState, ConsumerModuleState: consumerModuleState}
}
