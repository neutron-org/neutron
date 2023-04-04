package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/app/params"

	"github.com/neutron-org/neutron/app"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/x/feerefunder/types"
)

const (
	TestAddressNeutron         = "neutron13xvjxhkkxxhztcugr6weyt76eedj5ucpt4xluv"
	TestContractAddressJuno    = "juno10h0hc64jv006rr8qy0zhlu4jsxct8qwa0vtaleayh0ujz0zynf2s2r7v8q"
	TestContractAddressNeutron = "neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq"
)

func TestGenesisState_Validate(t *testing.T) {
	cfg := app.GetDefaultConfig()
	cfg.Seal()

	validRecvFee := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(0)))
	validAckFee := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(types.DefaultFees.AckFee.AmountOf(params.DefaultDenom).Int64()+1)))
	validTimeoutFee := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(types.DefaultFees.TimeoutFee.AmountOf(params.DefaultDenom).Int64()+1)))

	invalidRecvFee := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdk.NewInt(1)))

	validPacketId := types.NewPacketID("port", "channel-1", 64)

	for _, tc := range []struct {
		desc             string
		genState         *types.GenesisState
		valid            bool
		expectedErrorMsg string
	}{
		{
			desc:             "default is valid",
			genState:         types.DefaultGenesis(),
			valid:            true,
			expectedErrorMsg: "",
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
			valid:            true,
			expectedErrorMsg: "",
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
			valid:            false,
			expectedErrorMsg: "failed to parse the payer address",
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
			valid:            false,
			expectedErrorMsg: "is not a contract",
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
			valid:            false,
			expectedErrorMsg: "failed to parse the payer address",
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
			valid:            false,
			expectedErrorMsg: "port id",
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
			valid:            false,
			expectedErrorMsg: "channel id",
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
			valid:            false,
			expectedErrorMsg: "invalid fees",
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
				require.Contains(t, err.Error(), tc.expectedErrorMsg)
			}
		})
	}
}
