package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateDenom{}, "osmosis/coinfactory/create-denom", nil)
	cdc.RegisterConcrete(&MsgMint{}, "osmosis/coinfactory/mint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "osmosis/coinfactory/burn", nil)
	// cdc.RegisterConcrete(&MsgForceTransfer{}, "osmosis/coinfactory/force-transfer", nil)
	cdc.RegisterConcrete(&MsgChangeAdmin{}, "osmosis/coinfactory/change-admin", nil)
	cdc.RegisterConcrete(&MsgSetBeforeSendHook{}, "osmosis/coinfactory/set-beforesend-hook", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "osmosis/coinfactory/update-params", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateDenom{},
		&MsgMint{},
		&MsgBurn{},
		// &MsgForceTransfer{},
		&MsgChangeAdmin{},
		&MsgSetBeforeSendHook{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(amino)
	// TODO: RegisterCodec(authzcodec.Amino)

	amino.Seal()
}
