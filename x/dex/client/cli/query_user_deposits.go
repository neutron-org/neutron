package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v2/x/dex/types"
)

// TODO: add optional IncludePoolData arg
func CmdListUserDeposits() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-user-deposits [address]",
		Short:   "list all users deposits",
		Example: "list-user-deposits alice",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqAddress := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllUserDepositsRequest{
				Address: reqAddress,
			}

			res, err := queryClient.UserDepositsAll(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)

	return cmd
}
