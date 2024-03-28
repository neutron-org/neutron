package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDeposit{}, "dex/Deposit", nil)
	cdc.RegisterConcrete(&MsgWithdrawal{}, "dex/Withdrawal", nil)
	cdc.RegisterConcrete(&MsgPlaceLimitOrder{}, "dex/PlaceLimitOrder", nil)
	cdc.RegisterConcrete(&MsgWithdrawFilledLimitOrder{}, "dex/WithdrawFilledLimitOrder", nil)
	cdc.RegisterConcrete(&MsgCancelLimitOrder{}, "dex/CancelLimitOrder", nil)
	cdc.RegisterConcrete(&MsgMultiHopSwap{}, "dex/MultiHopSwap", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeposit{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWithdrawal{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPlaceLimitOrder{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWithdrawFilledLimitOrder{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelLimitOrder{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMultiHopSwap{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
