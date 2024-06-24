package main

import (
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/neutron-org/neutron/v4/wasmbinding/bindings"
	"github.com/spf13/cobra"
	"os"
)

func genWasmbindingSchemaCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "gen-wasmbinding-schema",
		Short: "Generates wasmbinding json schema for NeutronQuery and NeutronMsg",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {
			r := new(jsonschema.Reflector)
			if err := r.AddGoComments("github.com/neutron-org/neutron/v4", "./"); err != nil {
				// deal with error
				return err
			}
			querySchema := r.Reflect(&bindings.NeutronQuery{})

			queryJson, err := json.MarshalIndent(querySchema, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to create jsonschema: %w", err)
			}

			err = writeToFile(err, queryJson, "wasmbinding/schema/query.json")
			if err != nil {
				return err
			}

			msgSchema := r.Reflect(&bindings.NeutronMsg{})
			msgJson, err := json.MarshalIndent(msgSchema, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to create jsonschema: %w", err)
			}

			err = writeToFile(err, msgJson, "wasmbinding/schema/msg.json")
			if err != nil {
				return fmt.Errorf("failed to write to a file: %w", err)
			}

			fmt.Println("wasmbinding json schema generated successfully")

			return nil
		},
	}

	return txCmd
}

func writeToFile(err error, bts []byte, filepath string) error {
	file, err := os.Create(filepath)

	if err != nil {
		return fmt.Errorf("failed to create file with path=%s: %w", filepath, err)
	}
	// Close the file at the end
	defer file.Close()

	_, err = file.Write(bts)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
