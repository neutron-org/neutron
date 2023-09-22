package types_test

import (
	"testing"

	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				LimitOrderTrancheUserList: []*types.LimitOrderTrancheUser{
					{
						TrancheKey:  "0",
						Address:     "0",
						TradePairID: &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
					},
					{
						TrancheKey:  "1",
						Address:     "1",
						TradePairID: &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
					},
				},
				TickLiquidityList: []*types.TickLiquidity{
					{
						Liquidity: &types.TickLiquidity_LimitOrderTranche{
							LimitOrderTranche: &types.LimitOrderTranche{
								Key: &types.LimitOrderTrancheKey{
									TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
									TickIndexTakerToMaker: 0,
									TrancheKey:            "0",
								},
							},
						},
					},
					{
						Liquidity: &types.TickLiquidity_PoolReserves{
							PoolReserves: &types.PoolReserves{
								Key: &types.PoolReservesKey{
									TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
									TickIndexTakerToMaker: 0,
									Fee:                   0,
								},
							},
						},
					},
				},
				InactiveLimitOrderTrancheList: []*types.LimitOrderTranche{
					{
						Key: &types.LimitOrderTrancheKey{
							TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
							TickIndexTakerToMaker: 0,
							TrancheKey:            "0",
						},
					},
					{
						Key: &types.LimitOrderTrancheKey{
							TradePairID:           &types.TradePairID{TakerDenom: "TokenA", MakerDenom: "TokenB"},
							TickIndexTakerToMaker: 1,
							TrancheKey:            "1",
						},
					},
				},
				PoolMetadataList: []types.PoolMetadata{
					{
						ID: 0,
					},
					{
						ID: 1,
					},
				},
				PoolCount: 2,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated LimitOrderTrancheUser",
			genState: &types.GenesisState{
				LimitOrderTrancheUserList: []*types.LimitOrderTrancheUser{
					{
						TrancheKey:  "0",
						Address:     "0",
						TradePairID: &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
					},
					{
						TrancheKey:  "0",
						Address:     "0",
						TradePairID: &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated tickLiquidity",
			genState: &types.GenesisState{
				TickLiquidityList: []*types.TickLiquidity{
					{
						Liquidity: &types.TickLiquidity_LimitOrderTranche{
							LimitOrderTranche: &types.LimitOrderTranche{
								Key: &types.LimitOrderTrancheKey{
									TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
									TickIndexTakerToMaker: 0,
									TrancheKey:            "0",
								},
							},
						},
					},
					{
						Liquidity: &types.TickLiquidity_LimitOrderTranche{
							LimitOrderTranche: &types.LimitOrderTranche{
								Key: &types.LimitOrderTrancheKey{
									TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
									TickIndexTakerToMaker: 0,
									TrancheKey:            "0",
								},
							},
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated inactiveLimitOrderTranche",
			genState: &types.GenesisState{
				InactiveLimitOrderTrancheList: []*types.LimitOrderTranche{
					{
						Key: &types.LimitOrderTrancheKey{
							TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
							TickIndexTakerToMaker: 0,
							TrancheKey:            "0",
						},
					},
					{
						Key: &types.LimitOrderTrancheKey{
							TradePairID:           &types.TradePairID{MakerDenom: "TokenA", TakerDenom: "TokenB"},
							TickIndexTakerToMaker: 0,
							TrancheKey:            "0",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated poolMetadata",
			genState: &types.GenesisState{
				PoolMetadataList: []types.PoolMetadata{
					{
						ID: 0,
					},
					{
						ID: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid poolCount",
			genState: &types.GenesisState{
				PoolMetadataList: []types.PoolMetadata{
					{
						ID: 1,
					},
				},
				PoolCount: 0,
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
