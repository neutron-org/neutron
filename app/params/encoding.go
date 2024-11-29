package params

import (
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	neutrontx "github.com/neutron-org/neutron/v5/tx"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	reg, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	marshaler := codec.NewProtoCodec(reg)

	signingOptions, err := tx.NewDefaultSigningOptions()
	if err != nil {
		panic(err)
	}
	txCfg := tx.NewTxConfig(
		marshaler,
		tx.DefaultSignModes,
		neutrontx.NewSignModeEIP191Handler(aminojson.SignModeHandlerOptions{
			FileResolver: signingOptions.FileResolver,
			TypeResolver: signingOptions.TypeResolver,
		}))

	return EncodingConfig{
		InterfaceRegistry: reg,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
