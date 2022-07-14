package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
	"github.com/spf13/cobra"
)

func CmdInterchainAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "interchainaccounts [connection-id] [owner-account]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.InterchainAccountAddress(cmd.Context(), &types.QueryInterchainAccountAddressRequest{
				OwnerAddress: args[1],
				ConnectionId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
