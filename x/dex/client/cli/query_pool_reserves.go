package cli

import (
	"context"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func CmdListPoolReserves() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-pool-reserves [pair-id] [token-in]",
		Short:   "Query AllPoolReserves",
		Example: "list-pool-reserves tokenA<>tokenB tokenA",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqPairID := args[0]
			reqTokenIn := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryAllPoolReservesRequest{
				PairId:     reqPairID,
				TokenIn:    reqTokenIn,
				Pagination: pageReq,
			}

			res, err := queryClient.PoolReservesAll(cmd.Context(), params)
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

func CmdShowPoolReserves() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-pool-reserves [pair-id] [tick-index] [token-in] [fee]",
		Short:   "shows a PoolReserves",
		Example: "show-pool-reserves tokenA<>tokenB [-5] tokenA 1",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argPairID := args[0]
			if strings.HasPrefix(args[1], "[") && strings.HasSuffix(args[1], "]") {
				args[1] = strings.TrimPrefix(args[1], "[")
				args[1] = strings.TrimSuffix(args[1], "]")
			}
			argTickIndex := args[1]
			argTokenIn := args[2]
			argFee := args[3]

			argTrancheKeyInt, err := strconv.ParseUint(argFee, 10, 0)
			if err != nil {
				return err
			}

			argTickIndexInt, err := strconv.ParseInt(argTickIndex, 10, 0)
			if err != nil {
				return err
			}

			params := &types.QueryGetPoolReservesRequest{
				PairId:    argPairID,
				TickIndex: argTickIndexInt,
				TokenIn:   argTokenIn,
				Fee:       argTrancheKeyInt,
			}

			res, err := queryClient.PoolReserves(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
