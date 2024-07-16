package dex

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
)

var _ autocli.HasAutoCLIConfig = AppModule{}

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "neutron.dex.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query module params",
				},
				{
					RpcMethod:      "LimitOrderTrancheUser",
					Use:            "show-limit-order-tranche-user [address] [tranche-key] ?(--calc-withdraw)",
					Short:          "Shows a LimitOrderTrancheUser",
					Example:        "show-limit-order-tranche-user neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td TRANCHEKEY123",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}, {ProtoField: "tranche_key"}},
				},
				{
					RpcMethod: "LimitOrderTrancheUserAll",
					Use:       "list-limit-order-tranche-user",
					Short:     "Queries list of all LimitOrderTrancheUser",
				},
				{
					RpcMethod:      "LimitOrderTrancheAll",
					Use:            "list-limit-order-tranche [pair-id] [token-in]",
					Short:          "List all LimitOrderTranches",
					Example:        "list-limit-order-tranche tokenA<>tokenB tokenA",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "token_in"}},
				},
				{
					RpcMethod:      "LimitOrderTranche",
					Use:            "show-limit-order-tranche [pair-id] [tick-index] [token-in] [tranche-key]",
					Short:          "Shows a LimitOrderTranche",
					Example:        "show-limit-order-tranche tokenA<>tokenB [5] tokenA 0",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "tick_index"}, {ProtoField: "token_in"}, {ProtoField: "tranche_key"}},
				},
				{
					RpcMethod:      "LimitOrderTrancheUserAllByAddress",
					Use:            "list-user-limit-orders [address]",
					Short:          "List all users limit orders",
					Example:        "list-user-limit-orders alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod: "InactiveLimitOrderTrancheAll",
					Use:       "list-filled-limit-order-tranche",
					Short:     "List all InactiveLimitOrderTranche",
				},
				{
					RpcMethod:      "InactiveLimitOrderTranche",
					Use:            "show-filled-limit-order-tranche [pair-id] [token-in] [tick-index] [tranche-key]",
					Short:          "Shows a InactiveLimitOrderTranche",
					Example:        "show-filled limit-order-tranche tokenA<>tokenB tokenA [10] 0",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "token_in"}, {ProtoField: "tick_index"}, {ProtoField: "tranche_key"}},
				},
				{
					RpcMethod:      "UserDepositsAll",
					Use:            "list-user-deposits [address] ?(--include-pool-data)",
					Short:          "List all users deposits",
					Example:        "list-user-deposits alice --include-pool-data",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "Pool",
					Use:            "show-pool '[pair-id]' [tick-index] [fee]",
					Short:          "Show a pool. Make sure to wrap your pair-id in quotes otherwise the shell will interpret <> as a separator token",
					Example:        "show-pool 'tokenA<>tokenB' [-5] 1",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "tick_index"}, {ProtoField: "fee"}},
				},
				{
					RpcMethod:      "PoolByID",
					Use:            "show-pool-by-id [pool-id]",
					Short:          "Show a pool by poolID",
					Example:        "show-pool-by-id 5",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pool_id"}},
				},
				{
					RpcMethod: "PoolMetadataAll",
					Use:       "list-pool-metadata",
					Short:     "List all PoolMetadata",
				},
				{
					RpcMethod:      "PoolMetadata",
					Use:            "show-pool-metadata [pool-id]",
					Short:          "Show a PoolMetadata",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "TickLiquidityAll",
					Use:            "list-tick-liquidity [pair-id] [token-in]",
					Short:          "List all tickLiquidity",
					Example:        "list-tick-liquidity tokenA<>tokenB tokenA",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "token_in"}},
				},
				{
					RpcMethod:      "TickLiquidityAll",
					Use:            "list-tick-liquidity [pair-id] [token-in]",
					Short:          "List all tickLiquidity",
					Example:        "list-tick-liquidity tokenA<>tokenB tokenA",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "token_in"}},
				},
				{
					RpcMethod:      "PoolReservesAll",
					Use:            "list-pool-reserves [pair-id] [token-in]",
					Short:          "Query all PoolReserves",
					Example:        "list-pool-reserves tokenA<>tokenB tokenA",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "token_in"}},
				},
				{
					RpcMethod:      "PoolReserves",
					Use:            "show-pool-reserves [pair-id] [tick-index] [token-in] [fee]",
					Short:          "Shows PoolReserves",
					Example:        "show-pool-reserves tokenA<>tokenB [-5] tokenA 1",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pair_id"}, {ProtoField: "tick_index"}, {ProtoField: "token_in"}, {ProtoField: "fee"}},
				},
				// TODO: skip? wasn't in the original queries
				{
					RpcMethod: "EstimateMultiHopSwap",
					Use:       "estimate-multi-hop-swap [creator] [receiver] [amount-in] [exit-limit-price] (--routes 'route1,route2') (--pick_best_route)",
					Short:     "Estimates MultiHopSwap",
					Example:   "estimate-multi-hop-swap neutron1addr neutron2addr 100tokenA 0.15",
					//PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "creator"}, {ProtoField: "receiver"}, {ProtoField: "amount_in"}, {ProtoField: "exit_limit_price"},
				},
				// TODO: skip? wasn't in the original queries
				{
					RpcMethod: "EstimatePlaceLimitOrder",
					Use:       "estimate-place-limit-order [creator] [receiver] [token-in] [token-out] [tick-index-in-to-out] [amount-in] [order-type] [max-amount-out] (--timestamp)",
					Short:     "Estimates PlaceLimitOrder",
					Example:   "estimate-place-limit-order neutron1addr neutron2addr tokenA tokenB 10000 100 1 1721138602 ",
					//PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "creator"}, {ProtoField: "receiver"}, {ProtoField: "token_in"}, {ProtoField: "token_out"}, {ProtoField: "tick_index_in_to_out"}, {ProtoField: "amount_in"}, {ProtoField: "order_type"},{ProtoField: "max_amount_out"},
				},
			},
			EnhanceCustomCommand: true,
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "neutron.dex.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "CancelLimitOrder",
					Use:            "cancel-limit-order [tranche-key]",
					Short:          "Cancel limit order",
					Example:        "cancel-limit-order TRANCHEKEY123 --from alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "tranche_key"}},
				},
				{
					RpcMethod:      "Deposit",
					Use:            "deposit [receiver] [token-a] [token-b] [list of amount-0] [list of amount-1] [list of tick-index] [list of fees] [disable_autoswap] [fail_tx_on_bel]",
					Short:          "Make a deposit",
					Example:        "deposit alice tokenA tokenB 100,0 0,50 [-10,5] 1,1 false,false false,false --from alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "receiver"}, {ProtoField: "token_a"}, {ProtoField: "token_b"}, {ProtoField: "amounts_a"}, {ProtoField: "amounts_b"}, {ProtoField: "tick_indexes_a_to_b"}, {ProtoField: "fees"}, {ProtoField: "disable_autoswap"}, {ProtoField: "fail_tx_on_bel"}, {ProtoField: "fail_tx_on_bel"}},
				},
				{
					RpcMethod:      "Withdrawal",
					Use:            "withdrawal [receiver] [token-a] [token-b] [list of shares-to-remove] [list of tick-index] [list of fees]",
					Short:          "Withdraw pair",
					Example:        "withdrawal alice tokenA tokenB 100,50 [-10,5] 1,1 --from alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "receiver"}, {ProtoField: "token_a"}, {ProtoField: "token_b"}, {ProtoField: "shares_to_remove"}, {ProtoField: "tick_indexes_a_to_b"}, {ProtoField: "fees"}},
				},
				{
					RpcMethod: "",
				},
			},
		},
	}
}
