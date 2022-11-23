package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/app"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/x/feerefunder/types"
)

const TestAddressNeutron = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"
const TestContractAddressJuno = "juno10h0hc64jv006rr8qy0zhlu4jsxct8qwa0vtaleayh0ujz0zynf2s2r7v8q"
const TestContractAddressNeutron = "neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq"

func TestGenesisState_Validate(t *testing.T) {
	cfg := app.GetDefaultConfig()
	cfg.Seal()

	validRecvFee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))
	validAckFee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(types.DefaultFees.AckFee.AmountOf(sdk.DefaultBondDenom).Int64()+1)))
	validTimeoutFee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(types.DefaultFees.TimeoutFee.AmountOf(sdk.DefaultBondDenom).Int64()+1)))

	invalidRecvFee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
	invalidAckFee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(types.DefaultFees.AckFee.AmountOf(sdk.DefaultBondDenom).Int64()-1)))
	invalidTimeoutFee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(types.DefaultFees.TimeoutFee.AmountOf(sdk.DefaultBondDenom).Int64()-1)))

	validPacketId := types.NewPacketID("port", "channel-1", 64)

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
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: true,
		},
		{
			desc: "invalid payer address",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    "address",
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "payer is not a contract",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestAddressNeutron,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "payer is from a wrong chain",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressJuno,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "invalid port",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: types.NewPacketID("*", "channel", 64),
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "invalid channel",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: types.NewPacketID("port", "*", 64),
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "AckFee more than min",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     invalidAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "TimeoutFee more than min",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: invalidTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "Recv fee non-zero",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    invalidRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: false,
		},
		{
			desc: "Recv fee nil",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNeutron,
					PacketId: validPacketId,
					Fee: types.Fee{
						RecvFee:    nil,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: true,
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
