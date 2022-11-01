package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/spf13/cobra"
)

func CmdAllFailures() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "failures [address]",
		Short: "shows all failures or failures from specific contract address",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			address := ""
			if len(args) > 0 {
				address = args[0]
			}

			params := &types.QueryFailuresRequest{
				Address:    address,
				Pagination: pageReq,
			}

			res, err := queryClient.AllFailures(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
