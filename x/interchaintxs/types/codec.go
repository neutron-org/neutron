package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterInterchainAccount{}, "/lidofinance.interchainadapter.interchaintxs.v1.MsgRegisterInterchainAccount", nil)
	cdc.RegisterConcrete(&MsgSubmitTx{}, "/lidofinance.interchainadapter.interchaintxs.v1.MsgSubmitTx", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterInterchainAccount{},
		&MsgSubmitTx{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// TODO: as far as I understand, this should be removed, but currently
//  the testutil (which also probably should be removed) uses this
//  variable.
var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}
