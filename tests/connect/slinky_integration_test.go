package connect_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/connect/tests/integration/v2"
	marketmapmodule "github.com/skip-mev/connect/v2/x/marketmap"
	"github.com/skip-mev/connect/v2/x/oracle"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("neutron", "neutronpub")
	cfg.Seal()
}

var (
	image = ibc.DockerImage{
		Repository: "neutron-node",
		Version:    "latest",
		UidGid:     "1025:1025",
	}

	numValidators = 4
	numFullNodes  = 0
	noHostMount   = false
	gasAdjustment = 1.5

	oracleImage = ibc.DockerImage{
		Repository: "ghcr.io/skip-mev/connect-sidecar",
		Version:    "v2.1.0",
		UidGid:     "1000:1000",
	}
	encodingConfig = testutil.MakeTestEncodingConfig(
		bank.AppModuleBasic{},
		oracle.AppModuleBasic{},
		gov.AppModuleBasic{},
		auth.AppModuleBasic{},
		marketmapmodule.AppModuleBasic{},
	)

	VotingPeriod     = "10s"
	MaxDepositPeriod = "1s"

	defaultGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "consensus.params.abci.vote_extensions_enable_height",
			Value: "2",
		},
		{
			Key:   "consensus.params.block.max_gas",
			Value: "1000000000",
		},
		{
			Key:   "app_state.feemarket.params.enabled",
			Value: false,
		},
	}

	denom = "untrn"
	spec  = &interchaintest.ChainSpec{
		ChainName:     "connect",
		Name:          "connect",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       "latest",
		NoHostMount:   &noHostMount,
		ChainConfig: ibc.ChainConfig{
			EncodingConfig: &encodingConfig,
			Images: []ibc.DockerImage{
				image,
			},
			Type:           "cosmos",
			Name:           "connect",
			Denom:          denom,
			ChainID:        "chain-id-0",
			Bin:            "neutrond",
			Bech32Prefix:   "neutron",
			CoinType:       "118",
			GasAdjustment:  gasAdjustment,
			GasPrices:      fmt.Sprintf("0%s", denom),
			TrustingPeriod: "48h",
			NoHostMount:    noHostMount,
			ModifyGenesis:  cosmos.ModifyGenesis(defaultGenesisKV),
			SkipGenTx:      true,
		},
	}
)

func TestConnectOracleIntegration(t *testing.T) {
	baseSuite := integration.NewConnectIntegrationSuite(
		spec,
		oracleImage,
		integration.WithInterchainConstructor(integration.CCVInterchainConstructor),
		integration.WithChainConstructor(integration.CCVChainConstructor),
		integration.WithDenom(denom),
	)

	suite.Run(t, integration.NewConnectOracleIntegrationSuite(baseSuite))
}

func TestConnectCCVIntegration(t *testing.T) {
	s := integration.NewConnectCCVIntegrationSuite(
		spec,
		oracleImage,
		integration.WithInterchainConstructor(integration.CCVInterchainConstructor),
		integration.WithChainConstructor(integration.CCVChainConstructor),
		integration.WithDenom(denom),
	)

	suite.Run(t, s)
}
