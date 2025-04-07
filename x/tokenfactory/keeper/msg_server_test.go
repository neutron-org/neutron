package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/app/params"
	"github.com/neutron-org/neutron/v6/testutil"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/tokenfactory/keeper"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/keeper"
	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

const (
	testAddress = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"
	denom       = "factory/neutron1p87pglxer3rlqx5hafy2glszfdwhcg04qp6pj9/sun"
)

func TestMsgCreateDenomValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgCreateDenom
		expectedErr error
	}{
		{
			"empty sender",
			types.MsgCreateDenom{
				Sender:   "",
				Subdenom: "sun",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgCreateDenom{
				Sender:   "invalid_sender",
				Subdenom: "sun",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"long subdenom",
			types.MsgCreateDenom{
				Sender:   testutil.TestOwnerAddress,
				Subdenom: string(make([]byte, types.MaxSubdenomLength+1)),
			},
			types.ErrInvalidDenom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.CreateDenom(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgMintValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgMint
		expectedErr error
	}{
		{
			"empty sender",
			types.MsgMint{
				Sender:        "",
				Amount:        sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				MintToAddress: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgMint{
				Sender:        "invalid_sender",
				Amount:        sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				MintToAddress: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid mint_to_address",
			types.MsgMint{
				Sender:        testutil.TestOwnerAddress,
				MintToAddress: "invalid mint_to_address",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid coin denom",
			types.MsgMint{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  "{}!@#a",
					Amount: math.NewInt(100),
				},
				MintToAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"nil coin amount",
			types.MsgMint{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom: params.DefaultDenom,
				},
				MintToAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero coin amount",
			types.MsgMint{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  params.DefaultDenom,
					Amount: math.NewInt(0),
				},
				MintToAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"negative coin amount",
			types.MsgMint{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  params.DefaultDenom,
					Amount: math.NewInt(-100),
				},
				MintToAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.Mint(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgBurnValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgBurn
		expectedErr error
	}{
		{
			"empty sender",
			types.MsgBurn{
				Sender:          "",
				Amount:          sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				BurnFromAddress: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgBurn{
				Sender:          "invalid_sender",
				Amount:          sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				BurnFromAddress: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid burn_from_address",
			types.MsgBurn{
				Sender:          testutil.TestOwnerAddress,
				BurnFromAddress: "invalid burn_from_address",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid coin denom",
			types.MsgBurn{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  "{}!@#a",
					Amount: math.NewInt(100),
				},
				BurnFromAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"nil coin amount",
			types.MsgBurn{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom: params.DefaultDenom,
				},
				BurnFromAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"zero coin amount",
			types.MsgBurn{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  params.DefaultDenom,
					Amount: math.NewInt(0),
				},
				BurnFromAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"negative coin amount",
			types.MsgBurn{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  params.DefaultDenom,
					Amount: math.NewInt(-100),
				},
				BurnFromAddress: testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.Burn(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgForceTransferValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgForceTransfer
		expectedErr error
	}{
		{
			"empty sender",
			types.MsgForceTransfer{
				Sender:              "",
				Amount:              sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				TransferFromAddress: testAddress,
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgForceTransfer{
				Sender:              "invalid_sender",
				Amount:              sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				TransferFromAddress: testAddress,
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty address to transfer from",
			types.MsgForceTransfer{
				Sender:              testutil.TestOwnerAddress,
				Amount:              sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				TransferFromAddress: "",
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid address to transfer from",
			types.MsgForceTransfer{
				Sender:              testutil.TestOwnerAddress,
				Amount:              sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				TransferFromAddress: "invalid_address",
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty address to transfer to",
			types.MsgForceTransfer{
				Sender:              testutil.TestOwnerAddress,
				Amount:              sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				TransferFromAddress: testAddress,
				TransferToAddress:   "",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid address to transfer to",
			types.MsgForceTransfer{
				Sender:              testutil.TestOwnerAddress,
				Amount:              sdktypes.NewCoin(params.DefaultDenom, math.NewInt(100)),
				TransferFromAddress: testAddress,
				TransferToAddress:   "invalid_address",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid coin denom",
			types.MsgForceTransfer{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  "{}!@#a",
					Amount: math.NewInt(100),
				},
				TransferFromAddress: testAddress,
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"nil coin amount",
			types.MsgForceTransfer{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom: params.DefaultDenom,
				},
				TransferFromAddress: testAddress,
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
		{
			"negative coin amount",
			types.MsgForceTransfer{
				Sender: testutil.TestOwnerAddress,
				Amount: sdktypes.Coin{
					Denom:  params.DefaultDenom,
					Amount: math.NewInt(-100),
				},
				TransferFromAddress: testAddress,
				TransferToAddress:   testAddress,
			},
			sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.ForceTransfer(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgChangeAdminValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgChangeAdmin
		expectedErr error
	}{
		{
			"empty sender",
			types.MsgChangeAdmin{
				Sender:   "",
				Denom:    denom,
				NewAdmin: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgChangeAdmin{
				Sender:   "invalid_sender",
				Denom:    denom,
				NewAdmin: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty new admin",
			types.MsgChangeAdmin{
				Sender:   testutil.TestOwnerAddress,
				Denom:    denom,
				NewAdmin: "",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid new admin",
			types.MsgChangeAdmin{
				Sender:   testutil.TestOwnerAddress,
				Denom:    denom,
				NewAdmin: "invalid_address",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"not enough parts of denom",
			types.MsgChangeAdmin{
				Sender:   testutil.TestOwnerAddress,
				Denom:    "factory/sun",
				NewAdmin: testAddress,
			},
			types.ErrInvalidDenom,
		},
		{
			"incorrect denom prefix",
			types.MsgChangeAdmin{
				Sender:   testutil.TestOwnerAddress,
				Denom:    "bitcoin/factory/sun",
				NewAdmin: testAddress,
			},
			types.ErrInvalidDenom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.ChangeAdmin(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgSetDenomMetadataValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgSetDenomMetadata
		expectedErr string
	}{
		{
			"empty sender",
			types.MsgSetDenomMetadata{
				Sender: "",
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"Invalid sender address",
		},
		{
			"invalid sender",
			types.MsgSetDenomMetadata{
				Sender: "invalid_sender",
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"Invalid sender address",
		},
		{
			"empty metadata name",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "",
					Symbol:  denom,
				},
			},
			"name field cannot be blank",
		},
		{
			"empty metadata symbol",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "",
				},
			},
			"symbol field cannot be blank",
		},
		{
			"invalid metadata base denom",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    "{}&!",
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"invalid metadata base denom",
		},
		{
			"invalid metadata display denom",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: "{}&!",
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"invalid metadata display denom",
		},
		{
			"incorrect first denom unit",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    params.DefaultDenom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"metadata's first denomination unit must be the one with base denom",
		},
		{
			"incorrect exponent for the first denom unit",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 1,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			fmt.Sprintf("the exponent for base denomination unit %s must be 0", denom),
		},
		{
			"incorrect order of the denom units",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"denom units should be sorted asc by exponent",
		},
		{
			"duplicated denom units",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
						{
							Denom:    denom,
							Exponent: 1,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"duplicate denomination unit",
		},
		{
			"lack of the display denom in the denom units",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    denom,
					Display: params.DefaultDenom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"metadata must contain a denomination unit with display denom",
		},
		{
			"duplicated denom unit aliases",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{"sun", "sun"},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"duplicate denomination unit alias",
		},
		{
			"empty denom unit alias",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    denom,
							Exponent: 0,
							Aliases:  []string{""},
						},
					},
					Base:    denom,
					Display: denom,
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			fmt.Sprintf("alias for denom unit %s cannot be blank", denom),
		},
		{
			"not enough parts of metadata base denom",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "factory/sun",
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    "factory/sun",
					Display: "factory/sun",
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"not enough parts of denom",
		},
		{
			"incorrect metadata base denom prefix",
			types.MsgSetDenomMetadata{
				Sender: testutil.TestOwnerAddress,
				Metadata: banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "bitcoin/factory/sun",
							Exponent: 0,
							Aliases:  []string{"sun"},
						},
					},
					Base:    "bitcoin/factory/sun",
					Display: "bitcoin/factory/sun",
					Name:    "noname",
					Symbol:  "SUN",
				},
			},
			"denom prefix is incorrect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.SetDenomMetadata(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgSetBeforeSendHookValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name        string
		msg         types.MsgSetBeforeSendHook
		expectedErr error
	}{
		{
			"empty sender",
			types.MsgSetBeforeSendHook{
				Sender:       "",
				Denom:        denom,
				ContractAddr: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgSetBeforeSendHook{
				Sender:       "invalid_sender",
				Denom:        denom,
				ContractAddr: testAddress,
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid contract address",
			types.MsgSetBeforeSendHook{
				Sender:       testutil.TestOwnerAddress,
				Denom:        denom,
				ContractAddr: "invalid_address",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"not enough parts of denom",
			types.MsgSetBeforeSendHook{
				Sender:       testutil.TestOwnerAddress,
				Denom:        "factory/sun",
				ContractAddr: testAddress,
			},
			types.ErrInvalidDenom,
		},
		{
			"incorrect denom prefix",
			types.MsgSetBeforeSendHook{
				Sender:       testutil.TestOwnerAddress,
				Denom:        "bitcoin/factory/sun",
				ContractAddr: testAddress,
			},
			types.ErrInvalidDenom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.SetBeforeSendHook(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)

	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectedErr string
	}{
		{
			"empty authority",
			types.MsgUpdateParams{
				Authority: "",
			},
			"authority is invalid",
		},
		{
			"invalid authority",
			types.MsgUpdateParams{
				Authority: "invalid authority",
			},
			"authority is invalid",
		},
		{
			"empty fee_collector_address with denom_creation_fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					FeeCollectorAddress: "",
					DenomCreationFee:    sdktypes.NewCoins(sdktypes.NewCoin("untrn", math.OneInt())),
				},
			},
			"DenomCreationFee and FeeCollectorAddr must be both set or both unset",
		},
		{
			"fee_collector_address empty denom_creation_fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					FeeCollectorAddress: testAddress,
				},
			},
			"DenomCreationFee and FeeCollectorAddr must be both set or both unset",
		},
		{
			"invalid fee_collector_address",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					DenomCreationFee:    sdktypes.NewCoins(sdktypes.NewCoin("untrn", math.OneInt())),
					FeeCollectorAddress: "invalid fee_collector_address",
				},
			},
			"failed to validate FeeCollectorAddress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := k.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgUpdateParamsWhitelistedHooks(t *testing.T) {
	k, ctx := testkeeper.TokenFactoryKeeper(t, nil, nil, nil)

	tests := []struct {
		name   string
		params types.Params
		error  string
	}{
		{
			"success",
			types.Params{
				WhitelistedHooks: []*types.WhitelistedHook{{DenomCreator: testAddress, CodeID: 1}},
			},
			"",
		},
		{
			"success multiple ",
			types.Params{
				WhitelistedHooks: []*types.WhitelistedHook{
					{DenomCreator: testAddress, CodeID: 1},
					{DenomCreator: testAddress, CodeID: 2},
				},
			},
			"",
		},
		{
			"invalid denom creator",
			types.Params{
				WhitelistedHooks: []*types.WhitelistedHook{
					{DenomCreator: "bad_address", CodeID: 1},
				},
			},
			"invalid denom creator",
		},
		{
			"duplicate hooks",
			types.Params{
				WhitelistedHooks: []*types.WhitelistedHook{
					{DenomCreator: testAddress, CodeID: 1},
					{DenomCreator: testAddress, CodeID: 1},
				},
			},
			"duplicate whitelisted hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params:    tt.params,
			}
			resp, err := k.UpdateParams(ctx, msg)
			if len(tt.error) > 0 {
				require.ErrorContains(t, err, tt.error)
				require.Nil(t, resp)

			} else {
				require.NoError(t, err)
				newParams := k.GetParams(ctx)
				require.Equal(t, tt.params, newParams)
			}
		})
	}
}
