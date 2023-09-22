package cli

import (
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/utils/dcli"
	"github.com/neutron-org/neutron/x/epochs/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := dcli.QueryIndexCmd(types.ModuleName)
	dcli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdEpochInfos)
	dcli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdCurrentEpoch)

	return cmd
}

func GetCmdEpochInfos() (*dcli.QueryDescriptor, *types.QueryEpochsInfoRequest) {
	return &dcli.QueryDescriptor{
		Use:   "epoch-infos",
		Short: "Query running epoch infos.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}}`,
		QueryFnName: "EpochInfos",
	}, &types.QueryEpochsInfoRequest{}
}

func GetCmdCurrentEpoch() (*dcli.QueryDescriptor, *types.QueryCurrentEpochRequest) {
	return &dcli.QueryDescriptor{
		Use:   "current-epoch",
		Short: "Query current epoch by specified identifier.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} day`,
	}, &types.QueryCurrentEpochRequest{}
}
