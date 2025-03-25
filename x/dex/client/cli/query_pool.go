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

func CmdShowPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-pool '[pair-id]' [tick-index] [fee]",
		Short:   "shows a pool. Make sure to wrap your pair-id in quotes otherwise the shell will interpret <> as a separator token",
		Example: "show-pool 'tokenA<>tokenB' [-5] 1",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argPairID := args[0]
			if strings.HasPrefix(args[1], "[") && strings.HasSuffix(args[1], "]") {
				args[1] = strings.TrimPrefix(args[1], "[")
				args[1] = strings.TrimSuffix(args[1], "]")
			}
			argTickIndex := args[1]
			argFee := args[2]

			argFeeInt, err := strconv.ParseUint(argFee, 10, 0)
			if err != nil {
				return err
			}

			argTickIndexInt, err := strconv.ParseInt(argTickIndex, 10, 0)
			if err != nil {
				return err
			}

			params := &types.QueryPoolRequest{
				PairId:    argPairID,
				TickIndex: argTickIndexInt,
				Fee:       argFeeInt,
			}

			res, err := queryClient.Pool(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowPoolByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-pool-by-id [poolID]",
		Short:   "shows a pool by poolID",
		Example: "show-pool-by-id 5",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argPoolID, err := strconv.ParseUint(args[0], 10, 0)
			if err != nil {
				return err
			}

			params := &types.QueryPoolByIDRequest{
				PoolId: argPoolID,
			}

			res, err := queryClient.PoolByID(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
