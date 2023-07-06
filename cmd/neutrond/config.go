package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cosmos/cosmos-sdk/client"
	scconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

// This code is copied from the Juno implementation: https://github.com/CosmosContracts/juno/pull/601/files

type NeutronCustomClient struct {
	scconfig.ClientConfig
	Gas           string `mapstructure:"gas" json:"gas"`
	GasPrices     string `mapstructure:"gas-prices" json:"gas-prices"`
	GasAdjustment string `mapstructure:"gas-adjustment" json:"gas-adjustment"`

	Fees string `mapstructure:"fees" json:"fees"`
	// FeeGranter string `mapstructure:"fee-granter" json:"fee-granter"`
	// FeePayer   string `mapstructure:"fee-payer" json:"fee-payer"`

	Note string `mapstructure:"note" json:"note"`
}

// ConfigCmd returns a CLI command to interactively create an application CLI
// config file.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query an application CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	clientCtx := client.GetClientContextFromCmd(cmd)
	configPath := filepath.Join(clientCtx.HomeDir, "config")

	conf, err := getClientConfig(configPath, clientCtx.Viper)
	if err != nil {
		return fmt.Errorf("couldn't get client config: %w", err)
	}

	ncc := NeutronCustomClient{
		*conf,
		os.Getenv("NEUTROND_GAS"),
		os.Getenv("NEUTROND_GAS_PRICES"),
		os.Getenv("NEUTROND_GAS_ADJUSTMENT"),

		os.Getenv("NEUTROND_FEES"),
		// os.Getenv("NEUTROND_FEE_GRANTER"),
		// os.Getenv("NEUTROND_FEE_PAYER"),

		os.Getenv("NEUTROND_NOTE"),
	}

	switch len(args) {
	case 0:
		s, err := json.MarshalIndent(ncc, "", "\t")
		if err != nil {
			return fmt.Errorf("unable to marshal neutron custom client: %w", err)
		}

		cmd.Println(string(s))

	case 1:
		// it's a get
		key := args[0]

		switch key {
		case flags.FlagChainID:
			cmd.Println(conf.ChainID)
		case flags.FlagKeyringBackend:
			cmd.Println(conf.KeyringBackend)
		case tmcli.OutputFlag:
			cmd.Println(conf.Output)
		case flags.FlagNode:
			cmd.Println(conf.Node)
		case flags.FlagBroadcastMode:
			cmd.Println(conf.BroadcastMode)

		// Custom flags
		case flags.FlagGas:
			cmd.Println(ncc.Gas)
		case flags.FlagGasPrices:
			cmd.Println(ncc.GasPrices)
		case flags.FlagGasAdjustment:
			cmd.Println(ncc.GasAdjustment)
		case flags.FlagFees:
			cmd.Println(ncc.Fees)
		// TODO: add with newer version of cosmos-sdk
		// case flags.FlagFeeGranter:
		//	cmd.Println(ncc.FeeGranter)
		// case flags.FlagFeePayer:
		//	cmd.Println(ncc.FeePayer)
		case flags.FlagNote:
			cmd.Println(ncc.Note)
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
		}

	case 2:
		// it's set
		key, value := args[0], args[1]

		switch key {
		case flags.FlagChainID:
			ncc.ChainID = value
		case flags.FlagKeyringBackend:
			ncc.KeyringBackend = value
		case tmcli.OutputFlag:
			ncc.Output = value
		case flags.FlagNode:
			ncc.Node = value
		case flags.FlagBroadcastMode:
			ncc.BroadcastMode = value
		case flags.FlagGas:
			ncc.Gas = value
		case flags.FlagGasPrices:
			ncc.GasPrices = value
			ncc.Fees = "" // resets since we can only use 1 at a time
		case flags.FlagGasAdjustment:
			ncc.GasAdjustment = value
		case flags.FlagFees:
			ncc.Fees = value
			ncc.GasPrices = "" // resets since we can only use 1 at a time
			// TODO: add with newer version of cosmos-sdk
		// case flags.FlagFeeGranter:
		//	 ncc.FeeGranter = value
		// case flags.FlagFeePayer:
		//	 ncc.FeePayer = value
		case flags.FlagNote:
			ncc.Note = value
		default:
			return errUnknownConfigKey(key)
		}

		confFile := filepath.Join(configPath, "client.toml")
		if err := writeConfigToFile(confFile, &ncc); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}

	default:
		return fmt.Errorf("too many arguments: accepts between 0 and 2 args")
	}

	return nil
}

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                          ###
###############################################################################

# The network chain ID
chain-id = "{{ .ChainID }}"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "{{ .KeyringBackend }}"
# CLI output format (text|json)
output = "{{ .Output }}"
# <host>:<port> to Tendermint RPC interface for this chain
node = "{{ .Node }}"
# Transaction broadcasting mode (sync|async|block)
broadcast-mode = "{{ .BroadcastMode }}"

###############################################################################
###                          Neutron Tx Configuration                          ###
###############################################################################

# Amount of gas per transaction
gas = "{{ .Gas }}"
# Price per unit of gas (ex: 0.005untrn)
gas-prices = "{{ .GasPrices }}"
gas-adjustment = "{{ .GasAdjustment }}"

# Fees to use instead of set gas prices
fees = "{{ .Fees }}"

# Memo to include in your Transactions
note = "{{ .Note }}"
`

// TODO: add under fee line when cosmos-sdk version merged
// fee-granter = "{{ .FeeGranter }}"
// fee-payer = "{{ .FeePayer }}"

// writeConfigToFile parses defaultConfigTemplate, renders config using the template and writes it to
// configFilePath.
func writeConfigToFile(configFilePath string, config *NeutronCustomClient) error {
	var buffer bytes.Buffer

	tmpl := template.New("clientConfigFileTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return err
	}

	if err := configTemplate.Execute(&buffer, config); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), 0o600)
}

// getClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func getClientConfig(configPath string, v *viper.Viper) (*scconfig.ClientConfig, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(scconfig.ClientConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
