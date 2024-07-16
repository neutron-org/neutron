package cli

import (
	"fmt"
	"strconv"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/spf13/cobra"

	contractmanagertypes "github.com/neutron-org/neutron/v4/x/contractmanager/types"
)

// CmdFailureDetails returns the command handler for the failure's detailed error querying.
func CmdFailureDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "failure-details [address] [failure-id]",
		Short:   "Query the detailed error related to a contract's sudo call failure",
		Long:    "Query the detailed error related to a contract's sudo call failure based on contract's address and failure ID",
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf("%s query failure-details neutron1m0z0kk0qqug74n9u9ul23e28x5fszr628h20xwt6jywjpp64xn4qatgvm0 0", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]
			failureID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid failure ID %s: expected a uint64: %v", args[1], err)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := contractmanagertypes.NewQueryClient(clientCtx)
			if _, err = queryClient.AddressFailure(
				cmd.Context(),
				&contractmanagertypes.QueryFailuresRequest{Address: address, FailureId: failureID},
			); err != nil {
				return err
			}

			searchEvents := []string{
				fmt.Sprintf("%s.%s='%s'", wasmtypes.EventTypeSudo, wasmtypes.AttributeKeyContractAddr, address),
				fmt.Sprintf("%s.%s='%d'", wasmtypes.EventTypeSudo, contractmanagertypes.AttributeKeySudoFailureID, failureID),
			}
			// TODO: search events
			result, err := authtx.QueryTxsByEvents(clientCtx, 1, 1, strings.Join(searchEvents, ","), "") // only a single tx for a pair address+failure_id is expected
			if err != nil {
				return err
			}

			for _, tx := range result.Txs {
				for _, event := range tx.Events {
					if event.Type == wasmtypes.EventTypeSudo {
						for _, attr := range event.Attributes {
							if attr.Key == contractmanagertypes.AttributeKeySudoError {
								return clientCtx.PrintString(attr.Value)
							}
						}
					}
				}
			}
			return fmt.Errorf("detailed failure error message not found in node events")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
