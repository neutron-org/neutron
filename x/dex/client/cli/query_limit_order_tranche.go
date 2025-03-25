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

func CmdListLimitOrderTranche() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-limit-order-tranche [pair-id] [token-in]",
		Short:   "list all LimitOrderTranches",
		Example: "list-limit-order-tranche tokenA<>tokenB tokenA",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			argPairID := args[0]
			argTokenIn := args[1]

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllLimitOrderTrancheRequest{
				Pagination: pageReq,
				PairId:     argPairID,
				TokenIn:    argTokenIn,
			}

			res, err := queryClient.LimitOrderTrancheAll(context.Background(), params)
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

func CmdShowLimitOrderTranche() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-limit-order-tranche [pair-id] [tick-index] [token-in] [tranche-key]",
		Short:   "shows a LimitOrderTranche",
		Example: "show-limit-order-tranche tokenA<>tokenB [5] tokenA 0",
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
			argTrancheKey := args[3]

			argTickIndexInt, err := strconv.ParseInt(argTickIndex, 10, 0)
			if err != nil {
				return err
			}

			params := &types.QueryGetLimitOrderTrancheRequest{
				PairId:     argPairID,
				TickIndex:  argTickIndexInt,
				TokenIn:    argTokenIn,
				TrancheKey: argTrancheKey,
			}

			res, err := queryClient.LimitOrderTranche(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
