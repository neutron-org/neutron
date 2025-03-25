package cli

import (
	"context"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func CmdListInactiveLimitOrderTranche() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-filled-limit-order-tranche",
		Short: "list all InactiveLimitOrderTranche",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllInactiveLimitOrderTrancheRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.InactiveLimitOrderTrancheAll(context.Background(), params)
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

func CmdShowInactiveLimitOrderTranche() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-filled-limit-order-tranche [pair-id] [token-in] [tick-index] [tranche-key]",
		Short:   "shows a InactiveLimitOrderTranche",
		Example: "show-filled limit-order-tranche tokenA<>tokenB tokenA [10] 0",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argPairID := args[0]
			argTokenIn := args[1]

			if strings.HasPrefix(args[2], "[") && strings.HasSuffix(args[2], "]") {
				args[2] = strings.TrimPrefix(args[2], "[")
				args[2] = strings.TrimSuffix(args[2], "]")
			}
			argTickIndex, err := cast.ToInt64E(args[2])
			argTrancheKey := args[3]
			if err != nil {
				return err
			}

			params := &types.QueryGetInactiveLimitOrderTrancheRequest{
				PairId:     argPairID,
				TokenIn:    argTokenIn,
				TickIndex:  argTickIndex,
				TrancheKey: argTrancheKey,
			}

			res, err := queryClient.InactiveLimitOrderTranche(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
