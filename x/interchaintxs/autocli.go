package interchaintxs

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "neutron.interchaintxs.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query module params",
				},
				{
					RpcMethod: "InterchainAccountAddress",
					Use:       "interchain-account [owner-address] [connection-id] [interchain-account-id]",
					Short:     "Query the interchain account address for a specific combination of owner-address, connection-id and interchain-account-id",
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
