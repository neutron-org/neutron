package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	icqtypes "github.com/neutron-org/neutron/v7/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v7/x/state-verifier/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group harpoon queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryConsensusState())
	cmd.AddCommand(CmdVefifyStorageValues())

	return cmd
}

func CmdQueryConsensusState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consensus-state [height]",
		Short:   "returns saves consensus state by a particular height",
		Example: "consensus-state 12345",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			request := &types.QueryConsensusStateRequest{
				Height: height,
			}

			res, err := queryClient.QueryConsensusState(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdVefifyStorageValues() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify-storage-values [height] [storage-values-json-file]",
		Short:   "verifies storage values that saved in a file for a particular height",
		Example: "verify-storage-values 12345 values.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			storageValuesFile, err := os.ReadFile(args[1])
			if err != nil {
				return errors.Wrap(err, "failed to read storage values file")
			}

			var storageValues []*icqtypes.StorageValue
			if err := json.Unmarshal(storageValuesFile, &storageValues); err != nil {
				return errors.Wrap(err, "failed to unmarshal storage values file")
			}

			request := &types.QueryVerifyStateValuesRequest{
				Height:        height,
				StorageValues: storageValues,
			}

			res, err := queryClient.VerifyStateValues(context.Background(), request)
			if err != nil {
				return errors.Wrapf(err, "failed to verify storage values for height %d", height)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
