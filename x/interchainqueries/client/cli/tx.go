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

const (
	FlagTxFilter = "tx-filter"
	FlagKVKeys   = "kv-keys"
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
		Use:     "register-interchain-query [zone-id] [connection-id] [query-data] [update-period] [query_type]",
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
			updatePeriod, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse update-period: %w", err)
			}

			queryType := types.InterchainQueryType(args[4])
			if !queryType.IsValid() {
				return fmt.Errorf("invalid query type: must be %s or %s, got %s", types.InterchainQueryTypeKV, types.InterchainQueryTypeTX, queryType)
			}

			var (
				txFilter     string
				kvKeysString []string
				kvKeys       []*types.KVKey
			)

			if queryType.IsTX() {
				txFilter, err = cmd.Flags().GetString(FlagTxFilter)
				if err != nil {
					return err
				}
			}

			if queryType.IsKV() {
				kvKeysString, err = cmd.Flags().GetStringSlice(FlagKVKeys)
				if err != nil {
					return err
				}

				kvKeys = make([]*types.KVKey, 0, len(kvKeysString))
				for _, k := range kvKeysString {
					key, err := types.KVKeyFromString(k)
					if err != nil {
						return fmt.Errorf("failed to parse kv key from string: %w", err)
					}

					kvKeys = append(kvKeys, &key)
				}
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

	cmd.Flags().String(FlagTxFilter, "", `filter for ICQ transactions search: --tx-filter {"message.module": "bank"}`)
	cmd.Flags().StringSlice(FlagKVKeys, []string{}, "kv keys for transactions search: --kv-keys acc/01dadc1baa5cbc9e36485e690aa433fa33396e9b87")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func SubmitQueryResultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "submit-query-result [query-id] [result-file]",
		Short:   "Submit query result",
		Aliases: []string{"submit", "s"},
		Args:    cobra.ExactArgs(5),
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
