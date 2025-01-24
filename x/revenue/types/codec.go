package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "neutron.revenue.MsgUpdateParams", nil)

	cdc.RegisterInterface((*PaymentSchedule)(nil), nil)
	cdc.RegisterConcrete(&MonthlyPaymentSchedule{}, "neutron/MonthlyPaymentSchedule", nil)
	cdc.RegisterConcrete(&BlockBasedPaymentSchedule{}, "neutron/BlockBasedPaymentSchedule", nil)
	cdc.RegisterConcrete(&EmptyPaymentSchedule{}, "neutron/EmptyPaymentSchedule", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	registry.RegisterInterface(
		"neutron.revenue.PaymentSchedule",
		(*PaymentSchedule)(nil),
		&MonthlyPaymentSchedule{},
		&BlockBasedPaymentSchedule{},
		&EmptyPaymentSchedule{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
