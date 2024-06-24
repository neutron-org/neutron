package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v4/wasmbinding/bindings"
)

var (
	schemaFolder = "wasmbinding/schema/"
	packageName  = "github.com/neutron-org/neutron/v4"
	prefix       = ""
	indent       = "  "
	queryFile    = "query.json"
	msgFile      = "msg.json"
)

func genWasmbindingSchemaCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "gen-wasmbinding-schema",
		Short: "Generates wasmbinding json schema for NeutronQuery and NeutronMsg",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			r := new(jsonschema.Reflector)
			if err := r.AddGoComments(packageName, "./"); err != nil {
				return fmt.Errorf("failed to add comments: %w", err)
			}
			querySchema := r.Reflect(&bindings.NeutronQuery{})

			queryJSON, err := json.MarshalIndent(querySchema, prefix, indent)
			if err != nil {
				return fmt.Errorf("failed to create jsonschema: %w", err)
			}

			err = writeToFile(queryJSON, schemaFolder+queryFile)
			if err != nil {
				return err
			}

			msgSchema := r.Reflect(&bindings.NeutronMsg{})
			msgJSON, err := json.MarshalIndent(msgSchema, prefix, indent)
			if err != nil {
				return fmt.Errorf("failed to create jsonschema: %w", err)
			}

			err = writeToFile(msgJSON, schemaFolder+msgFile)
			if err != nil {
				return fmt.Errorf("failed to write to a file: %w", err)
			}

			fmt.Println("wasmbinding json schema generated successfully")

			return nil
		},
	}

	return txCmd
}

func writeToFile(bts []byte, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file with path=%s: %w", filepath, err)
	}

	defer file.Close()

	_, err = file.Write(bts)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
