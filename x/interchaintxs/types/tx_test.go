package types_test

import (
	"testing"

	"cosmossdk.io/math"
	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v3/app"
	feetypes "github.com/neutron-org/neutron/v3/x/feerefunder/types"
	"github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

const TestAddress = "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2"

func TestMsgRegisterInterchainAccountValidate(t *testing.T) {
	_ = app.GetDefaultConfig()

	tests := []struct {
		name        string
		malleate    func() sdktypes.HasValidateBasic
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.HasValidateBasic {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
				}
			},
			nil,
		},
		{
			"empty fromAddress",
			func() sdktypes.HasValidateBasic {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         "",
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid fromAddress",
			func() sdktypes.HasValidateBasic {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         "invalid address",
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty connection id",
			func() sdktypes.HasValidateBasic {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         TestAddress,
					ConnectionId:        "",
					InterchainAccountId: "1",
				}
			},
			types.ErrEmptyConnectionID,
		},
		{
			"empty interchain account",
			func() sdktypes.HasValidateBasic {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "",
				}
			},
			types.ErrEmptyInterchainAccountID,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgRegisterInterchainAccountGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}

func TestMsgSubmitTXValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.HasValidateBasic
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			nil,
		},
		{
			"invalid timeout",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 0,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			types.ErrInvalidTimeout,
		},
		{
			"empty connection id",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			types.ErrEmptyConnectionID,
		},
		{
			"empty interchain account id",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			types.ErrEmptyInterchainAccountID,
		},
		{
			"no messages",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs:                nil,
					Timeout:             1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			types.ErrNoMessages,
		},
		{
			"empty FromAddress",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         "",
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid FromAddress",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         "invalid_address",
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"non zero recv fee",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero ack fee",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(),
						TimeoutFee: sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
					},
				}
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero timeout fee",
			func() sdktypes.HasValidateBasic {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
					Timeout: 1,
					Fee: feetypes.Fee{
						RecvFee:    sdktypes.NewCoins(),
						AckFee:     sdktypes.NewCoins(sdktypes.NewCoin("denom", math.NewInt(100))),
						TimeoutFee: sdktypes.NewCoins(),
					},
				}
			},
			sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgSubmitTXGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}
