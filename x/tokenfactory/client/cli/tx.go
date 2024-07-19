package cli

import (
	"encoding/json"
	"fmt"
	"os"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/neutron-org/neutron/v4/x/tokenfactory/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewSetDenomMetadataCmd(),
	)

	return cmd
}

// NewSetDenomMetadataCmd broadcast MsgSetDenomMetadata msg
func NewSetDenomMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-denom-metadata [metadata-file] [flags]",
		Short: "Sets a bank metadata for a token denom. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			metadataBytes, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var metadata banktypes.Metadata
			if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
				return err
			}

			msg := types.NewMsgSetDenomMetadata(
				clientCtx.GetFromAddress().String(),
				metadata,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
