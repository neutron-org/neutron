package tokenfactory

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "osmosis.tokenfactory.v1beta1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query module params",
				},
				{
					RpcMethod:      "DenomAuthorityMetadata",
					Use:            "denom-authority-metadata [creator] [subdenom] [flags]",
					Short:          "Get the authority metadata for a specific denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "creator"}, {ProtoField: "subdenom"}},
				},
				{
					RpcMethod:      "DenomsFromCreator",
					Use:            "denoms-from-creator [creator-address] [flags]",
					Short:          "Returns a list of all tokens created by a specific creator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "creator"}},
				},
				{
					RpcMethod:      "BeforeSendHookAddress",
					Use:            "before-send-hook [creator] [subdenom] [flags]",
					Short:          "Get the before send hook for a specific denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "creator"}, {ProtoField: "subdenom"}},
				},
			},
			EnhanceCustomCommand: true,
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "osmosis.tokenfactory.v1beta1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
				{
					RpcMethod:      "CreateDenom",
					Use:            "create-denom [subdenom] [flags]",
					Short:          "Create a new denom from an account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subdenom"}},
				},
				{
					RpcMethod:      "Mint",
					Use:            "mint [amount] [flags]",
					Short:          "Mint a denom to an address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}},
				},
				{
					RpcMethod:      "Burn",
					Use:            "burn [amount] [flags]",
					Short:          "Burn tokens from an address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}},
				},
				{
					RpcMethod:      "ChangeAdmin",
					Use:            "change-admin [denom] [new-admin-address] [flags]",
					Short:          "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "denom"}, {ProtoField: "new_admin"}},
				},
				{
					RpcMethod:      "SetBeforeSendHook",
					Use:            "set-before-send-hook [denom] [contract-addr] [flags]",
					Short:          "Sets the before send hook for a factory-created denom. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "denom"}, {ProtoField: "contract_addr"}},
				},
				{
					RpcMethod:      "ForceTransfer",
					Use:            "force-transfer [amount] [transfer-from-address] [transfer-to-address] [flags]",
					Short:          "Force transfer tokens from one address to another address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}, {ProtoField: "transferFromAddress"}, {ProtoField: "transferToAddress"}},
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
