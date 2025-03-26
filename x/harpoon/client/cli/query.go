package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group harpoon queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQuerySubscribedContracts())

	return cmd
}

func CmdQuerySubscribedContracts() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscribed-contracts [hook-type]",
		Short:   "lists all contracts subscribed to given hook type",
		Example: "subscribed-contracts HOOK_TYPE_AFTER_VALIDATOR_CREATED",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QuerySubscribedContractsRequest{
				HookType: types.HookType(types.HookType_value[args[0]]),
			}

			res, err := queryClient.SubscribedContracts(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
