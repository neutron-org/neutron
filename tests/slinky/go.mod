module neutron/tests/slinky

go 1.22.5

toolchain go1.22.6

replace (
	cosmossdk.io/core => cosmossdk.io/core v0.11.0
	github.com/ChainSafe/go-schnorrkel => github.com/ChainSafe/go-schnorrkel v0.0.0-20200405005733-88cbf1b4c40d
	github.com/ChainSafe/go-schnorrkel/1 => github.com/ChainSafe/go-schnorrkel v1.0.0
	github.com/docker/distribution => github.com/docker/distribution v2.8.2+incompatible
	github.com/docker/docker => github.com/docker/docker v24.0.9+incompatible
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/vedhavyas/go-subkey => github.com/strangelove-ventures/go-subkey v1.0.7
)

require (
	github.com/cosmos/cosmos-sdk v0.50.9
	github.com/skip-mev/connect/tests/integration/v2 v2.0.0-20240919172831-1508062c5eb8
	github.com/skip-mev/connect/v2 v2.0.1
	github.com/strangelove-ventures/interchaintest/v8 v8.7.0
	github.com/stretchr/testify v1.9.0
)

replace github.com/cosmos/cosmos-sdk => github.com/neutron-org/cosmos-sdk v0.50.8-neutron
