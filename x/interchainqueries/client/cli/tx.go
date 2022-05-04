package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/types"
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

	cmd.AddCommand(RegisterInterchainQueryCmd())

	return cmd
}

func RegisterInterchainQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-interchain-query [zone-id] [connection-id] [query-type] [query-data] [update-period]",
		Short:   "Register an interchain query",
		Aliases: []string{"register", "r"},
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()
			zoneID := args[0]
			connectionID := args[1]
			queryType := args[2]
			queryData := args[3]
			updatePeriod, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse update-period: %w", err)
			}

			msg := types.MsgRegisterInterchainQuery{
				QueryData:    queryData,
				QueryType:    queryType,
				ZoneId:       zoneID,
				ConnectionId: connectionID,
				UpdatePeriod: updatePeriod,
				Sender:       sender.String(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
