package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/neutron-org/neutron/app"
)

func main() {
	config := app.GetDefaultConfig()
	config.Seal()

	rootCmd, _ := NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "NEUTRON", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
