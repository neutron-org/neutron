package keeper_test

// import (
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// var (
// 	defaultAddr  sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
// 	defaultCoins sdk.Coins      = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
// )

// func (suite *KeeperTestSuite) measureStakeGas(addr sdk.AccAddress, coins sdk.Coins, dur time.Duration) uint64 {
// 	// fundAccount outside of gas measurement
// 	suite.FundAcc(addr, coins)
// 	// start measuring gas
// 	alreadySpent := suite.Ctx.GasMeter().GasConsumed()
// 	_, err := suite.App.IncentivesKeeper.CreateStake(suite.Ctx, addr, coins, dur)
// 	suite.Require().NoError(err)
// 	newSpent := suite.Ctx.GasMeter().GasConsumed()
// 	spentNow := newSpent - alreadySpent
// 	return spentNow
// }

// func (suite *KeeperTestSuite) measureAvgAndMaxStakeGas(
// 	numIterations int,
// 	addr sdk.AccAddress,
// 	coinsFn func(int) sdk.Coins,
// 	durFn func(int) time.Duration,
// ) (avg uint64, maxGas uint64) {
// 	runningTotal := uint64(0)
// 	maxGas = uint64(0)
// 	for i := 1; i <= numIterations; i++ {
// 		stakeGas := suite.measureStakeGas(addr, coinsFn(i), durFn(i))
// 		runningTotal += stakeGas
// 		if stakeGas > maxGas {
// 			maxGas = stakeGas
// 			// fmt.Println(suite.Ctx.GasMeter().String())
// 		}
// 	}
// 	avg = runningTotal / uint64(numIterations)
// 	return avg, maxGas
// }

// // This maintains hard coded gas test vector changes,
// // so we can easily track changes
// func (suite *KeeperTestSuite) TestRepeatedStakeTokensGas() {
// 	suite.SetupTest()

// 	coinsFn := func(int) sdk.Coins { return defaultCoins }
// 	durFn := func(int) time.Duration { return time.Second }
// 	startAveragingAt := 1000
// 	totalNumStakes := 10000

// 	firstStakeGasAmount := suite.measureStakeGas(defaultAddr, defaultCoins, time.Second)
// 	suite.Assert().LessOrEqual(int(firstStakeGasAmount), 100000)

// 	for i := 1; i < startAveragingAt; i++ {
// 		suite.SetupStake(defaultAddr, defaultCoins)
// 	}
// 	avgGas, maxGas := suite.measureAvgAndMaxStakeGas(totalNumStakes-startAveragingAt, defaultAddr, coinsFn, durFn)
// 	fmt.Printf("test deets: total stakes created %d, begin average at %d\n", totalNumStakes, startAveragingAt)
// 	suite.Assert().LessOrEqual(int(avgGas), 100000, "average gas / stake")
// 	suite.Assert().LessOrEqual(int(maxGas), 100000, "max gas / stake")
// }

// func (suite *KeeperTestSuite) TestRepeatedStakeTokensDistinctDurationGas() {
// 	suite.SetupTest()

// 	coinsFn := func(int) sdk.Coins { return defaultCoins }
// 	durFn := func(i int) time.Duration { return time.Duration(i+1) * time.Second }
// 	totalNumStakes := 10000

// 	avgGas, maxGas := suite.measureAvgAndMaxStakeGas(totalNumStakes, defaultAddr, coinsFn, durFn)
// 	fmt.Printf("test deets: total stakes created %d\n", totalNumStakes)
// 	suite.Assert().LessOrEqual(int(avgGas), 150000, "average gas / stake")
// 	suite.Assert().LessOrEqual(int(maxGas), 300000, "max gas / stake")
// }
