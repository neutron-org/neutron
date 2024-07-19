package globalfee

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

func (a AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "gaia.globalfee.v1beta1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query module params",
					Long:      "Shows globalfee requirement: minimum_gas_prices, bypass_min_fee_msg_types, max_total_bypass_minFee_msg_gas_usage",
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
