package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/spf13/cobra"
)

func CmdListTickLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-tick-liquidity [pair-id] [token-in]",
		Short:   "list all tickLiquidity",
		Example: "list-tick-liquidity tokenA<>tokenB tokenA",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			argPairID := args[0]
			argTokenIn := args[1]

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllTickLiquidityRequest{
				PairID:     argPairID,
				TokenIn:    argTokenIn,
				Pagination: pageReq,
			}

			res, err := queryClient.TickLiquidityAll(context.Background(), params)
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
