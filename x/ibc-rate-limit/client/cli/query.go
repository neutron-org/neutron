package cli

import (
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group ibc-rate-limit queries under a subcommand
	//cmd := &cobra.Command{
	//	Use:                        types.ModuleName,
	//	Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
	//	DisableFlagParsing:         true,
	//	SuggestionsMinimumDistance: 2,
	//	RunE:                       client.ValidateCmd,
	//}

	//cmd.AddCommand(
	//	GetParams(),
	//)

	return nil
}

// GetParams returns the params for the module
func GetParams() *cobra.Command {
	//cmd := &cobra.Command{
	//	Use:   "params [flags]",
	//	Short: "Get the params for the x/ibc-rate-limit module",
	//	Args:  cobra.ExactArgs(0),
	//	RunE: func(cmd *cobra.Command, _ []string) error {
	//		clientCtx, err := client.GetClientQueryContext(cmd)
	//		if err != nil {
	//			return err
	//		}
	//		queryClient := types.NewQueryClient(clientCtx)
	//
	//		res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
	//		if err != nil {
	//			return err
	//		}
	//
	//		return clientCtx.PrintProto(res)
	//	},
	//}
	//
	//flags.AddQueryFlagsToCmd(cmd)

	return nil
}