package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/neutron-org/neutron/v5/x/tokenfactory/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group tokenfactory queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetParams(),
		GetCmdDenomAuthorityMetadata(),
		GetCmdDenomsFromCreator(),
		GetCmdBeforeSendHook(),
	)

	return cmd
}

// GetParams returns the params for the module
func GetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params [flags]",
		Short: "Get the params for the x/tokenfactory module",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdDenomAuthorityMetadata returns the authority metadata for a queried denom
func GetCmdDenomAuthorityMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-authority-metadata [denom] [flags]",
		Short: "Get the authority metadata for a specific denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			denom := strings.Split(args[0], "/")

			if len(denom) != 3 {
				return fmt.Errorf("invalid denom format, expected format: factory/[creator]/[subdenom]")
			}

			res, err := queryClient.DenomAuthorityMetadata(cmd.Context(), &types.QueryDenomAuthorityMetadataRequest{
				Creator:  denom[1],
				Subdenom: denom[2],
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

// GetCmdDenomsFromCreator a command to get a list of all tokens created by a specific creator address
func GetCmdDenomsFromCreator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denoms-from-creator [creator address] [flags]",
		Short: "Returns a list of all tokens created by a specific creator address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DenomsFromCreator(cmd.Context(), &types.QueryDenomsFromCreatorRequest{
				Creator: args[0],
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

// GetCmdBeforeSendHook returns the BeforeSendHook for a queried denom
func GetCmdBeforeSendHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "before-send-hook [denom] [flags]",
		Short: "Get the before send hook for a specific denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]
			creator, subdenom, err := types.DeconstructDenom(denom)
			if err != nil {
				return err
			}

			res, err := queryClient.BeforeSendHookAddress(cmd.Context(), &types.QueryBeforeSendHookAddressRequest{
				Creator:  creator,
				Subdenom: subdenom,
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
