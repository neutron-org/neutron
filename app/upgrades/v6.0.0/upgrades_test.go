package v600_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"

	"cosmossdk.io/math"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	types3 "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	consumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/app/params"
	v600 "github.com/neutron-org/neutron/v6/app/upgrades/v6.0.0"

	"github.com/neutron-org/neutron/v6/testutil"
)

//go:embed validators/staking
var Vals embed.FS

func commitBlock(chain *ibctesting.TestChain, res *abci.ResponseFinalizeBlock) {
	_, err := chain.App.Commit()
	require.NoError(chain.TB, err)

	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// val set changes returned from previous block get applied to the next validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain, chain.Vals, res.ValidatorUpdates)

	// increment the proposer priority of validators
	chain.Vals.IncrementProposerPriority(1)

	// increment the current header
	chain.CurrentHeader = cmtproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.Vals.Proposer.Address,
	}
}

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
}

func (suite *UpgradeTestSuite) TopUpWallet(ctx sdk.Context, sender, contractAddress sdk.AccAddress) {
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000_000_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func ExpectedVals() (map[string]ed25519.PubKey, error) {
	vals := make(map[string]ed25519.PubKey)
	err := fs.WalkDir(Vals, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := Vals.ReadFile(path)
		if err != nil {
			return err
		}
		skval := v600.StakingValidator{}
		err = json.Unmarshal(data, &skval)
		if err != nil {
			return err
		}
		vals[skval.Valoper] = skval.PK
		return nil
	})
	return vals, err
}

// Basic test to catch obvious errors
// A more comprehensive and complex test is implemented as TS scripts.
func (suite *UpgradeTestSuite) TestUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	_, addr, _ := bech32.DecodeAndConvert(v600.MainDAOContractAddress)
	suite.TopUpWallet(ctx, senderAddress, addr)

	// CCV SET
	// create a few ccv validators in the keeper
	const ccvNumber = 10
	ccvVals := map[string]types3.PubKey{}
	updates := []abci.ValidatorUpdate{}
	for i := 0; i < ccvNumber; i++ {
		ccvAddr := keeper.RandomAccountAddress(t)
		addr, err := bech32.ConvertAndEncode("neutronvaloper", ccvAddr)
		require.NoError(t, err)
		pk := ed25519.GenPrivKey().PubKey()
		ccvVals[addr] = pk
		anyPK, err := types2.NewAnyWithValue(pk)
		require.NoError(t, err)
		app.ConsumerKeeper.SetCCValidator(ctx, consumertypes.CrossChainValidator{
			Address:  ccvAddr,
			Power:    int64(100),
			Pubkey:   anyPK,
			OptedOut: false,
		})
		tmpk, err := codec.ToCmtProtoPublicKey(pk)
		require.NoError(t, err)
		valSetUpdate := abci.ValidatorUpdate{
			PubKey: tmpk,
			Power:  100,
		}
		updates = append(updates, valSetUpdate)
	}

	res, err := suite.ChainA.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             suite.ChainA.CurrentHeader.Height,
		Time:               suite.ChainA.CurrentHeader.GetTime(),
		NextValidatorsHash: suite.ChainA.NextVals.Hash(),
	})
	require.NoError(suite.ChainA.TB, err)
	res.ValidatorUpdates = updates
	// add ccv validators to consensus layer
	commitBlock(suite.ChainA, res)
	// new ctx with updated header
	ctx = suite.ChainA.GetContext().WithChainID("neutron-1")
	// CCV SET ^^^^^^^^^^^^^^^^

	// app initialized with predefined val set, we need to take it into account
	vals, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Equal(t, len(vals), 4)

	initialBonded, err := app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.BondedPoolName).String(),
		Denom:   params.DefaultDenom,
	})
	require.NoError(t, err)

	// DEICS
	err = v600.DeICS(ctx, *app.StakingKeeper, app.ConsumerKeeper, app.BankKeeper)
	require.NoError(t, err)

	expectedVals, err := ExpectedVals()
	require.NoError(t, err)

	newVals, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Equal(t, len(newVals), len(vals)+len(expectedVals)+ccvNumber)
	countStaking := 0
	countCCV := 0
	for _, newVal := range newVals {
		if _, ok := expectedVals[newVal.OperatorAddress]; ok {
			countStaking++
			require.Equal(t, newVal.Status, types.Unbonded)
			require.Equal(t, newVal.Tokens, math.NewInt(v600.SovereignSelfStake))
		}
		if _, ok := ccvVals[newVal.OperatorAddress]; ok {
			countCCV++
			require.Equal(t, newVal.Status, types.Bonded)
			require.Equal(t, newVal.Tokens, math.NewInt(v600.ICSSelfStake))
		}
	}
	// all expected new vals in the set
	require.Equal(t, countStaking, len(expectedVals))

	// all ccv vals in the set
	require.Equal(t, countCCV, len(ccvVals))

	bondedBalanceAfter, err := app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.BondedPoolName).String(),
		Denom:   params.DefaultDenom,
	})
	require.NoError(t, err)
	// ICS set adds stake to bonded pool
	require.Equal(t, bondedBalanceAfter.Balance.Amount.Sub(initialBonded.Balance.Amount), math.NewInt(v600.ICSSelfStake).MulRaw(int64(ccvNumber)))

	nonbondedBalanceAfter, err := app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.NotBondedPoolName).String(),
		Denom:   params.DefaultDenom,
	})
	require.NoError(t, err)
	// Sovereign set adds stake to nonbonded pool
	require.Equal(t, nonbondedBalanceAfter.Balance.Amount, math.NewInt(v600.SovereignSelfStake).MulRaw(int64(len(expectedVals))))

	err = v600.SetupRevenue(ctx, *app.RevenueKeeper, app.BankKeeper)
	require.NoError(t, err)

	resp, err := app.RevenueKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, resp.TwapWindow, int64(900))

	// TEST STAKING ENDBLOCKER and valset update
	// the tricky part is - we have valset of 4 initially, and we must to modify staking params to execute staking endblocker
	p, err := app.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	p.MaxValidators += 4
	err = app.StakingKeeper.SetParams(ctx, p)
	require.NoError(t, err)

	suite.ChainA.NextBlock()
	ctx = suite.ChainA.GetContext().WithChainID("neutron-1")
	bondedBalanceAfter, err = app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.BondedPoolName).String(),
		Denom:   params.DefaultDenom,
	})
	require.NoError(t, err)
	// Sovereign self stake moves to bonded pool
	require.Equal(t, bondedBalanceAfter.Balance.Amount.Sub(initialBonded.Balance.Amount), math.NewInt(v600.SovereignSelfStake).MulRaw(int64(len(expectedVals))))

	nonbondedBalanceAfter, err = app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.NotBondedPoolName).String(),
		Denom:   params.DefaultDenom,
	})
	require.NoError(t, err)
	// ICS self stake moves to bonded pool
	require.Equal(t, nonbondedBalanceAfter.Balance.Amount, math.NewInt(v600.ICSSelfStake).MulRaw(int64(ccvNumber)))

	// check valset at height =
	// `upgrade + 1` (ChainA.Vals) only contains initial staking + ccv
	// `upgrade + 2` (ChainA.NextVals) only contains initial staking + sovereign
	require.Equal(t, len(ccvVals)+len(vals), len(suite.ChainA.Vals.Validators))
	require.Equal(t, len(expectedVals)+len(vals), len(suite.ChainA.NextVals.Validators))
	// initial vals in both sets
	for _, v := range vals {
		pk, err := v.ConsPubKey()
		require.NoError(t, err)
		require.True(t, suite.ChainA.Vals.HasAddress(pk.Address()))
		require.True(t, suite.ChainA.NextVals.HasAddress(pk.Address()))
	}
	for _, pk := range expectedVals {
		require.False(t, suite.ChainA.Vals.HasAddress(pk.Address()))
		require.True(t, suite.ChainA.NextVals.HasAddress(pk.Address()))
	}
	for _, pk := range ccvVals {
		require.True(t, suite.ChainA.Vals.HasAddress(pk.Address()))
		require.False(t, suite.ChainA.NextVals.HasAddress(pk.Address()))
	}

	// dynamicfees
	err = v600.SetupDynamicfees(ctx, app.DynamicFeesKeeper)
	require.NoError(t, err)

	dfkParams := app.DynamicFeesKeeper.GetParams(ctx)
	require.Equal(t, dfkParams.NtrnPrices, sdk.DecCoins{sdk.DecCoin{Denom: v600.DropNtrnDenom, Amount: math.LegacyOneDec()}})
}
