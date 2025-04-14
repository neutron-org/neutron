package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        revenuetypes.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", revenuetypes.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdFundTreasury())

	return cmd
}

func CmdFundTreasury() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-treasury [amount]",
		Short: "Fund the revenue-treasury module account. The amount's denom must match the reward asset denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress().String()
			amount, err := sdktypes.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := revenuetypes.MsgFundTreasury{
				Sender: sender,
				Amount: sdktypes.NewCoins(amount),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
