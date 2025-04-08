package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

func CmdInterchainAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interchain-account [owner-address] [connection-id] [interchain-account-id]",
		Short: "get the interchain account address for a specific combination of owner-address, connection-id and interchain-account-id",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.InterchainAccountAddress(cmd.Context(), &types.QueryInterchainAccountAddressRequest{
				OwnerAddress:        args[0],
				ConnectionId:        args[1],
				InterchainAccountId: args[2],
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
