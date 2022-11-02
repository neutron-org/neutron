package cli

import (
	"context"
	"fmt"
	"strconv"

	// "strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/x/fee/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group fee queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdFeeInfo())

	return cmd
}

func CmdFeeInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-info [port_id] [channel_id] [sequence]",
		Short: "queries fee info by port id, channel id and sequnce",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			sequence, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse sequence: %w", err)
			}

			res, err := queryClient.FeeInfo(context.Background(), &types.FeeInfoRequest{
				ChannelId: args[1],
				PortId:    args[0],
				Sequence:  sequence,
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
