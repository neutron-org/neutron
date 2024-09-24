package feemarket_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	marketmapmodule "github.com/skip-mev/connect/v2/x/marketmap"
	"github.com/skip-mev/connect/v2/x/oracle"
	"github.com/skip-mev/feemarket/tests/e2e"
	feemarketmodule "github.com/skip-mev/feemarket/x/feemarket"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("neutron", "neutronpub")
	cfg.Seal()
}

var (
	minBaseGasPrice = sdkmath.LegacyMustNewDecFromStr("0.001")
	baseGasPrice    = sdkmath.LegacyMustNewDecFromStr("0.01")

	image = ibc.DockerImage{
		Repository: "neutron-node",
		Version:    "latest",
		UidGid:     "1025:1025",
	}

	oracleImage = ibc.DockerImage{
		Repository: "ghcr.io/skip-mev/slinky-sidecar",
		Version:    "latest",
		UidGid:     "1000:1000",
	}

	numValidators = 4
	numFullNodes  = 0
	noHostMount   = false
	gasAdjustment = 2.0

	encodingConfig = testutil.MakeTestEncodingConfig(
		bank.AppModuleBasic{},
		oracle.AppModuleBasic{},
		gov.AppModuleBasic{},
		auth.AppModuleBasic{},
		feemarketmodule.AppModuleBasic{},
		marketmapmodule.AppModuleBasic{},
	)

	defaultGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "consensus.params.abci.vote_extensions_enable_height",
			Value: "2",
		},
		{
			Key:   "consensus.params.block.max_gas",
			Value: strconv.Itoa(int(feemarkettypes.DefaultMaxBlockUtilization)),
		},
		{
			Key: "app_state.feemarket.params",
			Value: feemarkettypes.Params{
				Alpha:               feemarkettypes.DefaultAlpha,
				Beta:                feemarkettypes.DefaultBeta,
				Gamma:               feemarkettypes.DefaultAIMDGamma,
				Delta:               feemarkettypes.DefaultDelta,
				MinBaseGasPrice:     minBaseGasPrice,
				MinLearningRate:     feemarkettypes.DefaultMinLearningRate,
				MaxLearningRate:     feemarkettypes.DefaultMaxLearningRate,
				MaxBlockUtilization: 15_000_000,
				Window:              feemarkettypes.DefaultWindow,
				FeeDenom:            denom,
				Enabled:             true,
				DistributeFees:      false,
			},
		},
		{
			Key: "app_state.feemarket.state",
			Value: feemarkettypes.State{
				BaseGasPrice: baseGasPrice,
				LearningRate: feemarkettypes.DefaultMaxLearningRate,
				Window:       make([]uint64, feemarkettypes.DefaultWindow),
				Index:        0,
			},
		},
	}

	denom = "untrn"
	spec  = &interchaintest.ChainSpec{
		ChainName:     "feemarket",
		Name:          "feemarket",
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
			Name:           "feemarket",
			Denom:          denom,
			ChainID:        "chain-id-feemarket",
			Bin:            "neutrond",
			Bech32Prefix:   "neutron",
			CoinType:       "118",
			GasAdjustment:  gasAdjustment,
			GasPrices:      fmt.Sprintf("10%s", denom),
			TrustingPeriod: "48h",
			NoHostMount:    noHostMount,
			ModifyGenesis:  cosmos.ModifyGenesis(defaultGenesisKV),
			SkipGenTx:      true,
		},
	}

	txCfg = e2e.TestTxConfig{
		SmallSendsNum:          1,
		LargeSendsNum:          325,
		TargetIncreaseGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0011"),
	}
)

func TestE2ETestSuite(t *testing.T) {
	s := e2e.NewIntegrationSuite(
		spec,
		oracleImage,
		txCfg,
		e2e.WithInterchainConstructor(e2e.CCVInterchainConstructor),
		e2e.WithChainConstructor(e2e.CCVChainConstructor),
		e2e.WithDenom(denom),
		e2e.WithGasPrices(strings.Join([]string{"0.0uatom"}, ",")),
	)
	suite.Run(t, s)
}
