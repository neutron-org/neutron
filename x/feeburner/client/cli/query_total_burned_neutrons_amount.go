package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

func CmdQueryTotalBurnedNeutronsAmount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-burned-neutrons-amount",
		Short: "shows total amount of burned neutrons",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalBurnedNeutronsAmount(context.Background(), &types.QueryTotalBurnedNeutronsAmountRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
