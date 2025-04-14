package cli

import (
	"encoding/json"
	"fmt"
	"os"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCreateDenomCmd(),
		NewMintCmd(),
		NewBurnCmd(),
		NewForceTransferCmd(),
		NewChangeAdminCmd(),
		NewSetBeforeSendHook(),
		NewSetDenomMetadataCmd(),
	)

	return cmd
}

// NewCreateDenomCmd broadcast MsgCreateDenom
func NewCreateDenomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-denom [subdenom] [flags]",
		Short: "create a new denom from an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateDenom(
				clientCtx.GetFromAddress().String(),
				args[0],
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMintCmd broadcast MsgMint
func NewMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [amount] [flags]",
		Short: "Mint a denom to an address. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgMint(
				clientCtx.GetFromAddress().String(),
				amount,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewBurnCmd broadcast MsgBurn
func NewBurnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [amount] [flags]",
		Short: "Burn tokens from an address. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgBurn(
				clientCtx.GetFromAddress().String(),
				amount,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewForceTransferCmd broadcast MsgForceTransfer
func NewForceTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "force-transfer [amount] [transfer-from-address] [transfer-to-address] [flags]",
		Short: "Force transfer tokens from one address to another address. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgForceTransfer(
				clientCtx.GetFromAddress().String(),
				amount,
				args[1],
				args[2],
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSetDenomMetadataCmd broadcast MsgSetDenomMetadata msg
func NewSetDenomMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-denom-metadata [metadata-file] [flags]",
		Short: "Sets a bank metadata for a token denom. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			metadataBytes, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var metadata banktypes.Metadata
			if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
				return err
			}

			msg := types.NewMsgSetDenomMetadata(
				clientCtx.GetFromAddress().String(),
				metadata,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewChangeAdminCmd broadcast MsgChangeAdmin
func NewChangeAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-admin [denom] [new-admin-address] [flags]",
		Short: "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeAdmin(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSetBeforeSendHook broadcast MsgSetBeforeSendHook
func NewSetBeforeSendHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-before-send-hook [denom] [contract-addr] [flags]",
		Short: "Sets the before send hook for a factory-created denom. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			msg := types.NewMsgSetBeforeSendHook(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
