package app

import (
	"github.com/cosmos/cosmos-sdk/std"

	ethcryptocodec "github.com/neutron-org/neutron/v5/x/crypto/codec"

	"github.com/neutron-org/neutron/v6/app/params"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ethcryptocodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ethcryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
