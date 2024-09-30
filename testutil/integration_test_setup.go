package testutil

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	consumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"

	"github.com/neutron-org/neutron/v5/testutil/consumer"

	"github.com/stretchr/testify/require"
)

func Setup(t *testing.T) ibctesting.TestingApp {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(
		senderPrivKey.PubKey().Address().Bytes(),
		senderPrivKey.PubKey(),
		0,
		0,
	)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	ibctesting.DefaultTestingAppInit = SetupTestingApp(tmtypes.TM2PB.ValidatorUpdates(valSet))

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, "neutron-1", balance)

	return app
}

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(
	t *testing.T,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	chainID string,
	balances ...banktypes.Balance,
) ibctesting.TestingApp {
	t.Helper()

	app, genesisState := ibctesting.DefaultTestingAppInit()

	// ensure baseapp has a chain-id set before running InitChain
	baseapp.SetChainID(chainID)(app.GetBaseApp())

	genesisState, err := GenesisStateWithValSet(
		app.AppCodec(),
		genesisState,
		valSet,
		genAccs,
		balances...)
	require.NoError(t, err)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	_, err = app.InitChain(
		&abci.RequestInitChain{
			ChainId:         chainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: sims.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               time.Now(),
		NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	return app
}

// GenesisStateWithValSet returns a new genesis state with the validator set
func GenesisStateWithValSet(
	codec codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction
	initValPowers := []abci.ValidatorUpdate{}

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey: %w", err)
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create new any: %w", err)
		}

		validator := stakingtypes.Validator{
			OperatorAddress: sdk.ValAddress(val.Address).String(),
			ConsensusPubkey: pkAny,
			Jailed:          false,
			Status:          stakingtypes.Bonded,
			Tokens:          bondAmt,
			DelegatorShares: math.LegacyOneDec(),
			Description:     stakingtypes.Description{},
			UnbondingHeight: int64(0),
			UnbondingTime:   time.Unix(0, 0).UTC(),
			Commission: stakingtypes.NewCommission(
				math.LegacyZeroDec(),
				math.LegacyZeroDec(),
				math.LegacyZeroDec(),
			),
			MinSelfDelegation: math.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(
			delegations,
			stakingtypes.NewDelegation(
				genAccs[0].GetAddress().String(),
				sdk.ValAddress(val.Address).String(),
				math.LegacyOneDec(),
			),
		)

		// add initial validator powers so consumer InitGenesis runs correctly
		pub, _ := val.ToProto()
		initValPowers = append(initValPowers, abci.ValidatorUpdate{
			Power:  val.VotingPower,
			PubKey: pub.PubKey,
		})
	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(
		stakingtypes.DefaultParams(),
		validators,
		delegations,
	)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	vals, err := tmtypes.PB2TM.ValidatorUpdates(initValPowers)
	if err != nil {
		panic("failed to get vals")
	}

	consumerGenesisState := consumer.CreateMinimalConsumerTestGenesis()
	consumerGenesisState.Provider.InitialValSet = initValPowers
	consumerGenesisState.Provider.ConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).
		Hash()
	consumerGenesisState.Params.Enabled = true
	genesisState[consumertypes.ModuleName] = codec.MustMarshalJSON(consumerGenesisState)

	return genesisState, nil
}

// type EmptyAppOptions struct {
//	servertypes.AppOptions
// }

// // Get implements AppOptions
// func (ao EmptyAppOptions) Get(_ string) interface{} {
//	return nil
// }

// var _ network.TestFixtureFactory = NewTestNetworkFixture

// func NewTestNetworkFixture() network.TestFixture {
//	dir, err := os.MkdirTemp("", "neutron")
//	if err != nil {
//		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
//	}
//	defer os.RemoveAll(dir)

//	encConfig := MakeEncodingConfig()

//	app := NewApp(
//		log.NewNopLogger(),
//		dbm.NewMemDB(),
//		nil,
//		true,
//		map[int64]bool{},
//		DefaultNodeHome,
//		5,
//		EmptyAppOptions{},
//		encConfig,
//		nil,
//	)

//	appCtr := func(val network.ValidatorI) servertypes.Application {
//		return NewApp(
//			val.GetCtx().Logger,
//			dbm.NewMemDB(),
//			nil,
//			true,
//			map[int64]bool{},
//			DefaultNodeHome,
//			5,
//			EmptyAppOptions{},
//			encConfig,
//			nil,
//			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
//			bam.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
//			bam.SetChainID(val.GetCtx().Viper.GetString(flags.FlagChainID)),
//		)
//	}

//	return network.TestFixture{
//		AppConstructor: appCtr,
//		GenesisState:   NewDefaultGenesisState(app.AppCodec()),
//		EncodingConfig: moduletestutil.TestEncodingConfig{
//			InterfaceRegistry: app.InterfaceRegistry(),
//			Codec:             app.AppCodec(),
//			TxConfig:          encConfig.TxConfig,
//			Amino:             app.LegacyAmino(),
//		},
//	}
// }
