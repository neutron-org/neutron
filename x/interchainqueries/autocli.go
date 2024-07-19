package interchainqueries

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "neutron.interchainqueries.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query module params",
				},
				{
					RpcMethod: "RegisteredQueries",
					Use:       "registered-queries",
					Short:     "Query all the interchain queries in the module",
				},
				{
					RpcMethod:      "RegisteredQuery",
					Use:            "registered-query",
					Short:          "Queries registered interchain query by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "query_id"}},
				},
				{
					RpcMethod: "RegisteredQueries",
					Use:       "registered-queries",
					Short:     "Query all the interchain queries in the module",
				},
				{
					RpcMethod:      "QueryResult",
					Use:            "query-result [query-id]",
					Short:          "Query result for registered query",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "query_id"}},
				},
				{
					RpcMethod:      "LastRemoteHeight",
					Use:            "query-last-remote-height [connection-id]",
					Short:          "Query last remote height by connection id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "connection_id"}},
				},
			},
			EnhanceCustomCommand: true,
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "neutron.interchainqueries.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "RemoveInterchainQuery",
					Use:            "remove-interchain-query [query-id]",
					Short:          "Remove interchain query",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "query_id"}},
				},
				{
					RpcMethod: "RegisterInterchainQuery",
					Skip:      true,
				},
				{
					RpcMethod: "UpdateInterchainQuery",
					Skip:      true,
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
