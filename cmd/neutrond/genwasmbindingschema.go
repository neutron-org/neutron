package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v4/wasmbinding/bindings"
)

const (
	schemaFolder       = "wasmbinding/schema/"
	packageName        = "github.com/neutron-org/neutron/v4"
	prefix             = ""
	indent             = "  "
	queryFile          = "query.json"
	msgFile            = "msg.json"
	queryResponsesFile = "query_responses.json"
	msgResponsesFile   = "msg_responses.json"
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

			// query
			if err := generateJsonToFile(r, &bindings.NeutronQuery{}, queryFile); err != nil {
				return fmt.Errorf("failed to generate query: %w", err)
			}
			// msg
			if err := generateJsonToFile(r, &bindings.NeutronMsg{}, msgFile); err != nil {
				return fmt.Errorf("failed to generate msg: %w", err)
			}
			// query responses
			if err := generateJsonToFile(r, &bindings.NeutronQueryResponse{}, queryResponsesFile); err != nil {
				return fmt.Errorf("failed to generate query response: %w", err)
			}
			// msg responses
			if err := generateJsonToFile(r, &bindings.NeutronMsgResponse{}, msgResponsesFile); err != nil {
				return fmt.Errorf("failed to generate query response: %w", err)
			}

			fmt.Println("wasmbinding json schema generated successfully")

			return nil
		},
	}

	return txCmd
}

func generateJsonToFile(r *jsonschema.Reflector, reflected any, filename string) error {
	querySchema := r.Reflect(reflected)
	queryJSON, err := json.MarshalIndent(querySchema, prefix, indent)
	if err != nil {
		return fmt.Errorf("failed to create jsonschema: %w", err)
	}
	if err := writeToFile(queryJSON, schemaFolder+filename); err != nil {
		return fmt.Errorf("failed to write json schema to a file: %w", err)
	}
	return err
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
