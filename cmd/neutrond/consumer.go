package main

import (
	"cosmossdk.io/errors"
	"encoding/json"
	"fmt"
	"github.com/cometbft/cometbft/p2p"
	"github.com/neutron-org/neutron/v3/testutil/consumer"

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

				valDirs := []string{"./val_a", "./val_b", "./val_c", "./val_d"}

				var initialValset []types1.ValidatorUpdate
				var peerIds []string
				for _, valDir := range valDirs {
					privValidator := pvm.LoadFilePVEmptyState(valDir, "")
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
					nodeKey, err := p2p.LoadNodeKey(valDir + "/config/node_key.json")
					if err != nil {
						return err
					}
					peerIds = append(peerIds, string(nodeKey.ID()))
				}

				vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValset)
				if err != nil {
					return errors.Wrap(err, "could not convert val updates to validator set")
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

	return txCmd
}

//func GetConsumerSectionCmd(defaultNodeHome string) *cobra.Command {
//	genesisMutator := NewDefaultGenesisIO()
//
//	txCmd := &cobra.Command{
//		Use:                        "save-validator-key",
//		Short:                      "ONLY FOR TESTING PURPOSES! Gets public val key from home dir",
//		SuggestionsMinimumDistance: 2,
//		RunE: func(cmd *cobra.Command, _ []string) error {
//			return genesisMutator.AlterConsumerModuleState(cmd, func(state *GenesisData, _ map[string]json.RawMessage) error {
//				clientCtx := client.GetClientContextFromCmd(cmd)
//				serverCtx := server.GetServerContextFromCmd(cmd)
//				cfg := serverCtx.Config
//				cfg.SetRoot(clientCtx.HomeDir)
//				privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
//				pk, err := privValidator.GetPubKey()
//				if err != nil {
//					return nil
//				}
//				sdkPublicKey, err := cryptocodec.FromCmtPubKeyInterface(pk)
//				filePV := pvm.LoadFilePVEmptyState("todo", "")
//
//				res := base64.StdEncoding.EncodeToString(sdkPublicKey.Bytes())
//				fmt.Printf("public key in base64 format: %s\n", res)
//
//				return nil
//			})
//		},
//	}
//
//	txCmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
//	flags.AddQueryFlagsToCmd(txCmd)
//
//	return txCmd
//}

//func extractValidatorKey(config *config.Config) (crypto.PublicKey, error) {
//	privValidator := pvm.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
//	pk, err := privValidator.GetPubKey()
//	if err != nil {
//		return crypto.PublicKey{}, err
//	}
//	sdkPublicKey, err := cryptocodec.FromTmPubKeyInterface(pk)
//	if err != nil {
//		return crypto.PublicKey{}, err
//	}
//	tmProtoPublicKey, err := cryptocodec.ToTmProtoPublicKey(sdkPublicKey)
//	if err != nil {
//		return crypto.PublicKey{}, err
//	}
//	return tmProtoPublicKey, nil
//}

type GenesisMutator interface {
	AlterConsumerModuleState(cmd *cobra.Command, callback func(state *GenesisData, appState map[string]json.RawMessage) error) error
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
