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
			querySchema := r.Reflect(&bindings.NeutronQuery{})
			queryJSON, err := json.MarshalIndent(querySchema, prefix, indent)
			if err != nil {
				return fmt.Errorf("failed to create query jsonschema: %w", err)
			}
			if err := writeToFile(queryJSON, schemaFolder+queryFile); err != nil {
				return err
			}

			// msg
			msgSchema := r.Reflect(&bindings.NeutronMsg{})
			msgJSON, err := json.MarshalIndent(msgSchema, prefix, indent)
			if err != nil {
				return fmt.Errorf("failed to create msg jsonschema: %w", err)
			}
			if err := writeToFile(msgJSON, schemaFolder+msgFile); err != nil {
				return err
			}

			// query responses
			queryResponsesSchema := r.Reflect(&bindings.NeutronQueryResponse{})
			queryResponsesJSON, err := json.MarshalIndent(queryResponsesSchema, prefix, indent)
			if err != nil {
				return fmt.Errorf("failed to create query responses jsonschema: %w", err)
			}
			if err := writeToFile(queryResponsesJSON, schemaFolder+queryResponsesFile); err != nil {
				return err
			}

			// msg responses
			msgResponsesSchema := r.Reflect(&bindings.NeutronMsgResponse{})
			msgResponsesJSON, err := json.MarshalIndent(msgResponsesSchema, prefix, indent)
			if err != nil {
				return fmt.Errorf("failed to create msg responses jsonschema: %w", err)
			}
			if err := writeToFile(msgResponsesJSON, schemaFolder+msgResponsesFile); err != nil {
				return err
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
