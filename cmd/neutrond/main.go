package main

import (
	"os"

	"github.com/neutron-org/neutron/v5/app/config"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/neutron-org/neutron/v5/app"
)

func main() {
	config := config.GetDefaultConfig()
	config.Seal()

	rootCmd, _ := NewRootCmd()

	rootCmd.AddCommand(AddConsumerSectionCmd(app.DefaultNodeHome))

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
