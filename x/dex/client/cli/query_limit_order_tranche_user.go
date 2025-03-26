package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func CmdListLimitOrderTrancheUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-limit-order-tranche-user",
		Short: "list all LimitOrderTrancheUser",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllLimitOrderTrancheUserRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.LimitOrderTrancheUserAll(context.Background(), params)
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

func CmdShowLimitOrderTrancheUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-limit-order-tranche-user [address] [tranche-key] ?(--calc-withdraw)",
		Short:   "shows a LimitOrderTrancheUser",
		Example: "show-limit-order-tranche-user neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td  TRANCHEKEY123",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			calcWithdraw, err := cmd.Flags().GetBool(FlagCalcWithdraw)
			if err != nil {
				return err
			}
			params := &types.QueryGetLimitOrderTrancheUserRequest{
				Address:                args[0],
				TrancheKey:             args[1],
				CalcWithdrawableShares: calcWithdraw,
			}

			res, err := queryClient.LimitOrderTrancheUser(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetCalcWithdrawableAmount())

	return cmd
}
