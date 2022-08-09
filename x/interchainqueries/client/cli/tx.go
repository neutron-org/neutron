package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
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
	cmd.AddCommand(SubmitQueryResultCmd())

	return cmd
}

func RegisterInterchainQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-interchain-query [zone-id] [connection-id] [update-period] [query_type] [query_data]",
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
			updatePeriod, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse update-period: %w", err)
			}

			queryType := types.InterchainQueryType(args[3])
			if !queryType.IsValid() {
				return fmt.Errorf("invalid query type: must be %s or %s, got %s", types.InterchainQueryTypeKV, types.InterchainQueryTypeTX, queryType)
			}

			queryData := args[4]

			var (
				txFilter string
				kvKeys   []*types.KVKey
			)

			switch queryType {
			case types.InterchainQueryTypeKV:
				kvKeys, err = types.KVKeysFromString(queryData)
				if err != nil {
					return fmt.Errorf("failed to parse KV keys from string: %w", err)
				}
			case types.InterchainQueryTypeTX:
				txFilter = queryData
			}

			msg := types.MsgRegisterInterchainQuery{
				TransactionsFilter: txFilter,
				Keys:               kvKeys,
				QueryType:          string(queryType),
				ZoneId:             zoneID,
				ConnectionId:       connectionID,
				UpdatePeriod:       updatePeriod,
				Sender:             sender.String(),
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

func SubmitQueryResultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "submit-query-result [query-id] [result-file]",
		Short:   "Submit query result",
		Aliases: []string{"submit", "s"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress()
			queryID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse query id: %w", err)
			}

			resultFile := args[1]

			result, err := ioutil.ReadFile(resultFile)
			if err != nil {
				return fmt.Errorf("failed to read query result file: %w", err)
			}

			msg := types.MsgSubmitQueryResult{QueryId: queryID, Sender: string(sender)}
			if err := json.Unmarshal(result, &msg.Result); err != nil {
				return fmt.Errorf("failed to unmarshal query result: %w", err)
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
