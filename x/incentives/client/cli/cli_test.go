package cli_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/utils"
	"github.com/neutron-org/neutron/utils/dcli"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/client/cli"
	"github.com/neutron-org/neutron/x/incentives/types"
)

var testAddresses = utils.CreateRandomAccounts(3)

// Queries ////////////////////////////////////////////////////////////////////

func TestGetCmdGetModuleStatus(t *testing.T) {
	desc, _ := cli.GetCmdGetModuleStatus()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetModuleStatusRequest]{
		"basic test": {
			ExpectedQuery: &types.GetModuleStatusRequest{},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdGetGaugeByID(t *testing.T) {
	desc, _ := cli.GetCmdGetGaugeByID()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetGaugeByIDRequest]{
		"basic test": {
			Cmd: "1", ExpectedQuery: &types.GetGaugeByIDRequest{Id: 1},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdGauges(t *testing.T) {
	desc, _ := cli.GetCmdGauges()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetGaugesRequest]{
		"test ACTIVE_UPCOMING": {
			Cmd: "ACTIVE_UPCOMING TokenA",
			ExpectedQuery: &types.GetGaugesRequest{
				Status: types.GaugeStatus_ACTIVE_UPCOMING,
				Denom:  "TokenA",
			},
		},
		"test UPCOMING": {
			Cmd: "UPCOMING TokenA",
			ExpectedQuery: &types.GetGaugesRequest{
				Status: types.GaugeStatus_UPCOMING,
				Denom:  "TokenA",
			},
		},
		"test FINISHED": {
			Cmd: "FINISHED TokenA",
			ExpectedQuery: &types.GetGaugesRequest{
				Status: types.GaugeStatus_FINISHED,
				Denom:  "TokenA",
			},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdGetStakeByID(t *testing.T) {
	desc, _ := cli.GetCmdGetStakeByID()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetStakeByIDRequest]{
		"basic test": {
			Cmd: "1", ExpectedQuery: &types.GetStakeByIDRequest{StakeId: 1},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdStakes(t *testing.T) {
	desc, _ := cli.GetCmdStakes()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetStakesRequest]{
		"basic test": {
			Cmd: fmt.Sprintf("%s", testAddresses[0]),
			ExpectedQuery: &types.GetStakesRequest{
				Owner: testAddresses[0].String(),
			},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdFutureRewardEstimate(t *testing.T) {
	desc, _ := cli.GetCmdGetFutureRewardEstimate()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetFutureRewardEstimateRequest]{
		"basic test": {
			Cmd: fmt.Sprintf("%s [1,2,3] 1000", testAddresses[0]),
			ExpectedQuery: &types.GetFutureRewardEstimateRequest{
				Owner:     testAddresses[0].String(),
				StakeIds:  []uint64{1, 2, 3},
				NumEpochs: 1000,
			},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetAccountHistory(t *testing.T) {
	desc, _ := cli.GetCmdGetAccountHistory()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetAccountHistoryRequest]{
		"basic test": {
			Cmd: fmt.Sprintf("%s", testAddresses[0]),
			ExpectedQuery: &types.GetAccountHistoryRequest{
				Account: testAddresses[0].String(),
			},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

func TestGetGaugeQualifyingValue(t *testing.T) {
	desc, _ := cli.GetCmdGaugeQualifyingValue()
	tcs := map[string]dcli.QueryCliTestCase[*types.GetGaugeQualifyingValueRequest]{
		"basic test": {
			Cmd: "1",
			ExpectedQuery: &types.GetGaugeQualifyingValueRequest{
				Id: 1,
			},
		},
	}
	dcli.RunQueryTestCases(t, desc, tcs)
}

// TXS ////////////////////////////////////////////////////////////////////////

func TestNewCreateGaugeCmd(t *testing.T) {
	testTime := time.Unix(1681505514, 0).UTC()
	desc, _ := cli.NewCreateGaugeCmd()
	tcs := map[string]dcli.TxCliTestCase[*types.MsgCreateGauge]{
		"basic test": {
			Cmd: fmt.Sprintf(
				"TokenA TokenB 0 100 100TokenA,100TokenB 50 0 --from %s",
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgCreateGauge{
				IsPerpetual: false,
				Owner:       testAddresses[0].String(),
				DistributeTo: types.QueryCondition{
					PairID:    &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"},
					StartTick: 0,
					EndTick:   100,
				},
				Coins: sdk.NewCoins(
					sdk.NewCoin("TokenA", math.NewInt(100)),
					sdk.NewCoin("TokenB", math.NewInt(100)),
				),
				StartTime:         time.Unix(0, 0).UTC(),
				NumEpochsPaidOver: 50,
				PricingTick:       0,
			},
		},
		"tests with time (RFC3339)": {
			Cmd: fmt.Sprintf(
				"TokenA TokenB [-20] 20 100TokenA,100TokenB 50 0 --start-time %s --from %s",
				testTime.Format(time.RFC3339),
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgCreateGauge{
				IsPerpetual: false,
				Owner:       testAddresses[0].String(),
				DistributeTo: types.QueryCondition{
					PairID:    &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"},
					StartTick: -20,
					EndTick:   20,
				},
				Coins: sdk.NewCoins(
					sdk.NewCoin("TokenA", math.NewInt(100)),
					sdk.NewCoin("TokenB", math.NewInt(100)),
				),
				StartTime:         testTime,
				NumEpochsPaidOver: 50,
				PricingTick:       0,
			},
		},
		"tests with time (unix int)": {
			Cmd: fmt.Sprintf(
				"TokenA TokenB [-20] 20 100TokenA,100TokenB 50 0 --start-time %d --from %s",
				testTime.Unix(),
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgCreateGauge{
				IsPerpetual: false,
				Owner:       testAddresses[0].String(),
				DistributeTo: types.QueryCondition{
					PairID:    &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"},
					StartTick: -20,
					EndTick:   20,
				},
				Coins: sdk.NewCoins(
					sdk.NewCoin("TokenA", math.NewInt(100)),
					sdk.NewCoin("TokenB", math.NewInt(100)),
				),
				StartTime:         testTime,
				NumEpochsPaidOver: 50,
				PricingTick:       0,
			},
		},
		"tests with perpetual": {
			Cmd: fmt.Sprintf(
				"TokenA TokenB [-20] 20 100TokenA,100TokenB 50 0 --perpetual --from %s",
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgCreateGauge{
				IsPerpetual: true,
				Owner:       testAddresses[0].String(),
				DistributeTo: types.QueryCondition{
					PairID:    &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"},
					StartTick: -20,
					EndTick:   20,
				},
				Coins: sdk.NewCoins(
					sdk.NewCoin("TokenA", math.NewInt(100)),
					sdk.NewCoin("TokenB", math.NewInt(100)),
				),
				StartTime:         time.Unix(0, 0).UTC(),
				NumEpochsPaidOver: 1,
				PricingTick:       0,
			},
		},
	}
	dcli.RunTxTestCases(t, desc, tcs)
}

func TestNewAddToGaugeCmd(t *testing.T) {
	desc, _ := cli.NewAddToGaugeCmd()
	tcs := map[string]dcli.TxCliTestCase[*types.MsgAddToGauge]{
		"basic test": {
			Cmd: fmt.Sprintf("1 1000TokenA --from %s", testAddresses[0]),
			ExpectedMsg: &types.MsgAddToGauge{
				Owner:   testAddresses[0].String(),
				GaugeId: 1,
				Rewards: sdk.NewCoins(sdk.NewCoin("TokenA", math.NewInt(1000))),
			},
		},
		"multiple tokens": {
			Cmd: fmt.Sprintf("1 1000TokenA,1TokenZ --from %s", testAddresses[0]),
			ExpectedMsg: &types.MsgAddToGauge{
				Owner:   testAddresses[0].String(),
				GaugeId: 1,
				Rewards: sdk.NewCoins(
					sdk.NewCoin("TokenA", math.NewInt(1000)),
					sdk.NewCoin("TokenZ", math.NewInt(1)),
				),
			},
		},
	}
	dcli.RunTxTestCases(t, desc, tcs)
}

func TestNewStakeCmd(t *testing.T) {
	desc, _ := cli.NewStakeCmd()
	tcs := map[string]dcli.TxCliTestCase[*types.MsgStake]{
		"basic test": {
			Cmd: fmt.Sprintf("1000TokenA --from %s", testAddresses[0]),
			ExpectedMsg: &types.MsgStake{
				Owner: testAddresses[0].String(),
				Coins: sdk.NewCoins(sdk.NewCoin("TokenA", math.NewInt(1000))),
			},
		},
		"multiple tokens": {
			Cmd: fmt.Sprintf("1000TokenA,1TokenZ --from %s", testAddresses[0]),
			ExpectedMsg: &types.MsgStake{
				Owner: testAddresses[0].String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("TokenA", math.NewInt(1000)),
					sdk.NewCoin("TokenZ", math.NewInt(1)),
				),
			},
		},
		"tokenized share test": {
			Cmd: fmt.Sprintf(
				"1000DualityPoolShares-tokenA-tokenB-t123-f30 --from %s",
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgStake{
				Owner: testAddresses[0].String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("DualityPoolShares-tokenA-tokenB-t123-f30", math.NewInt(1000)),
				),
			},
		},
		"tokenized share negative tick index test": {
			Cmd: fmt.Sprintf(
				"1000DualityPoolShares-tokenA-tokenB-t-123-f30 --from %s",
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgStake{
				Owner: testAddresses[0].String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("DualityPoolShares-tokenA-tokenB-t-123-f30", math.NewInt(1000)),
				),
			},
		},
		"multiple tokenized shares": {
			Cmd: fmt.Sprintf(
				"1000DualityPoolShares-tokenA-tokenB-t-123-f30,1DualityPoolShares-tokenA-tokenB-t-124-f30 --from %s",
				testAddresses[0],
			),
			ExpectedMsg: &types.MsgStake{
				Owner: testAddresses[0].String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("DualityPoolShares-tokenA-tokenB-t-123-f30", math.NewInt(1000)),
					sdk.NewCoin("DualityPoolShares-tokenA-tokenB-t-124-f30", math.NewInt(1)),
				),
			},
		},
	}
	dcli.RunTxTestCases(t, desc, tcs)
}

func TestNewUnstakeCmd(t *testing.T) {
	desc, _ := cli.NewUnstakeCmd()
	tcs := map[string]dcli.TxCliTestCase[*types.MsgUnstake]{
		"basic test": {
			Cmd: fmt.Sprintf("--from %s", testAddresses[0]),
			ExpectedMsg: &types.MsgUnstake{
				Owner:    testAddresses[0].String(),
				Unstakes: []*types.MsgUnstake_UnstakeDescriptor{},
			},
		},
		"with coins": {
			Cmd: fmt.Sprintf("1:10TokenA 10:10TokenA,10TokenC --from %s", testAddresses[0]),
			ExpectedMsg: &types.MsgUnstake{
				Owner: testAddresses[0].String(),
				Unstakes: []*types.MsgUnstake_UnstakeDescriptor{
					{
						ID: 1,
						Coins: sdk.NewCoins(
							sdk.NewCoin("TokenA", math.NewInt(10)),
						),
					}, {
						ID: 10,
						Coins: sdk.NewCoins(
							sdk.NewCoin("TokenA", math.NewInt(10)),
							sdk.NewCoin("TokenC", math.NewInt(10)),
						),
					},
				},
			},
		},
	}
	dcli.RunTxTestCases(t, desc, tcs)
}
