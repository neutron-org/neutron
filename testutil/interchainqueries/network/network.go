package network

import (
	"fmt"
	"testing"
	"time"

	"github.com/neutron-org/neutron/v4/testutil"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	db "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/neutron-org/neutron/v4/app/params"

	"github.com/neutron-org/neutron/v4/app"
	"github.com/neutron-org/neutron/v4/testutil/consumer"
)

type (
	Network = network.Network
	Config  = network.Config
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	memoryDB := db.NewMemDB()

	// TODO: move to depinject
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used
	chainID := "chain-" + tmrand.NewRand().Str(6)
	tempHome := testutil.TestHomeDir(chainID)
	tempApp := app.New(
		log.NewNopLogger(),
		memoryDB,
		nil,
		false,
		map[int64]bool{},
		tempHome,
		0,
		sims.NewAppOptionsWithFlagHome(tempHome),
		nil,
	)
	encoding := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Marshaler:         tempApp.AppCodec(),
		TxConfig:          tempApp.GetTxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}
	// app doesn't have this module, but we need it for test setup, which uses MsgCreateValidator
	tempApp.BasicModuleManager[stakingtypes.ModuleName] = staking.AppModule{}
	tempApp.BasicModuleManager.RegisterInterfaces(encoding.InterfaceRegistry)

	return network.Config{
		Codec:             encoding.Marshaler,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			err := consumer.ModifyConsumerGenesis(val.(network.Validator))
			if err != nil {
				panic(err)
			}

			err = testutil.ModifyGenesisClearGenTxs(val.(network.Validator))
			if err != nil {
				panic(err)
			}

			return app.New(
				val.GetCtx().Logger, db.NewMemDB(), nil, true, map[int64]bool{}, val.GetCtx().Config.RootDir, 0,
				sims.EmptyAppOptions{},
				nil,
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID(chainID),
			)
		},
		GenesisState:  tempApp.BasicModuleManager.DefaultGenesis(encoding.Marshaler),
		TimeoutCommit: 2 * time.Second,
		ChainID:       chainID,
		// Some changes are introduced to make the tests run as if neutron is a standalone chain.
		// This will only work if NumValidators is set to 1.
		NumValidators:   1,
		BondDenom:       params.DefaultDenom,
		MinGasPrices:    fmt.Sprintf("0.000006%s", params.DefaultDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}
}
