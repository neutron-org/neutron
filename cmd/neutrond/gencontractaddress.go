package main

import (
	"fmt"
	"strconv"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/spf13/cobra"
)

func genContractAddressCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "generate-contract-address [instance_id] [code_id]",
		Short: "Generates contract address for the given instance id and code id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse instance_id: %w", err)
			}

			codeID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse code_id: %w", err)
			}

			contractAddress := wasmkeeper.BuildContractAddressClassic(codeID, instanceID)

			fmt.Println(contractAddress.String())

			return nil
		},
	}

	return txCmd
}
