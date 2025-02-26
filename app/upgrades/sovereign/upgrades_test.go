package v600_test

import (
	"cosmossdk.io/math"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	consumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	"github.com/neutron-org/neutron/v5/app/params"
	v600 "github.com/neutron-org/neutron/v5/app/upgrades/sovereign"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"testing"

	"github.com/neutron-org/neutron/v5/testutil"
)

//go:embed validators/staking
var Vals embed.FS

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
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func ExpectedVals() (map[string]struct{}, error) {
	vals := make(map[string]struct{})
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
		vals[skval.Valoper] = struct{}{}
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

	// create a few ccv validators in the keeper
	const ccvNumber = 10
	ccvVals := map[string]struct{}{}
	for i := 0; i < ccvNumber; i++ {
		ccvAddr := keeper.RandomAccountAddress(t)
		addr, err := bech32.ConvertAndEncode("neutronvaloper", ccvAddr)
		require.NoError(t, err)
		ccvVals[addr] = struct{}{}
		pk, err := types2.NewAnyWithValue(ed25519.GenPrivKey().PubKey())
		require.NoError(t, err)
		app.ConsumerKeeper.SetCCValidator(ctx, consumertypes.CrossChainValidator{
			Address:  ccvAddr,
			Power:    int64(i),
			Pubkey:   pk,
			OptedOut: false,
		})
	}

	// app initialized with predefined val set, we need to take it into account
	vals, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.Equal(t, len(vals), 4)

	bondedBalanceBefore, err := app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.BondedPoolName).String(),
		Denom:   "untrn",
	})
	require.NoError(t, err)

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
			require.Equal(t, newVal.Tokens, math.NewInt(v600.ICSValoperSelfStake))
		}
	}
	// all expected new vals in the set
	require.Equal(t, countStaking, len(expectedVals))

	// all ccv vals in the set
	require.Equal(t, countCCV, len(ccvVals))

	bondedBalanceAfter, err := app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.BondedPoolName).String(),
		Denom:   "untrn",
	})
	require.NoError(t, err)
	// ICS set adds stake to bonded pool
	require.Equal(t, bondedBalanceAfter.Balance.Amount.Sub(bondedBalanceBefore.Balance.Amount), math.NewInt(v600.ICSValoperSelfStake).MulRaw(int64(ccvNumber)))

	nonbondedBalanceAfter, err := app.BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: authtypes.NewModuleAddress(types.NotBondedPoolName).String(),
		Denom:   "untrn",
	})
	require.NoError(t, err)
	// Sovereign set adds stake to nonbonded pool
	require.Equal(t, nonbondedBalanceAfter.Balance.Amount, math.NewInt(v600.SovereignSelfStake).MulRaw(int64(len(expectedVals))))

	err = v600.SetupRevenue(ctx, *app.RevenueKeeper)
	require.NoError(t, err)

	resp, err := app.RevenueKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, resp.TwapWindow, int64(900))

}
