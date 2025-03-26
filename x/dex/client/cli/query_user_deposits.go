package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func CmdListUserDeposits() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-user-deposits [address] ?(--include-pool-data)",
		Short:   "list all users deposits",
		Example: "list-user-deposits alice --include-pool-data",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqAddress := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			includePoolData, err := cmd.Flags().GetBool(FlagIncludePoolData)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryAllUserDepositsRequest{
				Address:         reqAddress,
				IncludePoolData: includePoolData,
				Pagination:      pageReq,
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
	cmd.Flags().AddFlagSet(FlagSetIncludePoolData())

	return cmd
}
