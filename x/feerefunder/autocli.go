package feerefunder

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "neutron.feerefunder.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query module params",
				},
				{
					RpcMethod:      "FeeInfo",
					Use:            "fee-info [port_id] [channel_id] [sequence]",
					Short:          "Queries fee info by port id, channel id and sequence",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "port_id"}, {ProtoField: "channel_id"}, {ProtoField: "sequence"}},
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
