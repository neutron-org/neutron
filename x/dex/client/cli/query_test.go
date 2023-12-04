package cli

// import (
// 	"testing"

// 	"github.com/cosmos/cosmos-sdk/testutil/cli"
// 	"github.com/cosmos/cosmos-sdk/testutil/network"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/neutron-org/neutron/v2/app"
// 	dexclient "github.com/neutron-org/neutron/v2/x/dex/client/cli"
// 	"github.com/neutron-org/neutron/v2/x/dex/types"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// )

// type QueryTestSuite struct {
// 	suite.Suite

// 	cfg     network.Config
// 	network *network.Network

// 	addr1 sdk.AccAddress
// 	addr2 sdk.AccAddress
// }

// func (s *QueryTestSuite) SetupSuite() {
// 	var err error
// 	s.T().Log("setting up integration test suite")

// 	s.cfg = network.DefaultConfig(app.NewTestNetworkFixture)
// 	s.cfg.NumValidators = 1

// 	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
// 	s.Require().NoError(err)

// 	_, err = s.network.WaitForHeight(2)
// 	s.Require().NoError(err)
// }

// func TestQueryTestSuite(t *testing.T) {
// 	suite.Run(t, new(QueryTestSuite))
// }

// var testAddress = sdk.AccAddress([]byte("testAddr"))

// var limitOrderTrancheList = []types.TickLiquidity{
// 	{
// 		Liquidity: &types.TickLiquidity_LimitOrderTranche{
// 			LimitOrderTranche: &types.LimitOrderTranche{
// 				PairID: &types.PairID{
// 					Token0: "TokenA",
// 					Token1: "TokenB",
// 				},
// 				TokenIn:          "TokenB",
// 				TickIndex:        1,
// 				TrancheKey:       "0",
// 				ReservesTokenIn:  math.NewInt(10),
// 				ReservesTokenOut: math.ZeroInt(),
// 				TotalTokenIn:     math.NewInt(10),
// 				TotalTokenOut:    math.ZeroInt(),
// 			},
// 		},
// 	},
// 	{
// 		Liquidity: &types.TickLiquidity_LimitOrderTranche{
// 			LimitOrderTranche: &types.LimitOrderTranche{
// 				PairID: &types.PairID{
// 					Token0: "TokenA",
// 					Token1: "TokenB",
// 				},
// 				TokenIn:          "TokenB",
// 				TickIndex:        2,
// 				TrancheKey:       "1",
// 				ReservesTokenIn:  math.NewInt(10),
// 				ReservesTokenOut: math.ZeroInt(),
// 				TotalTokenIn:     math.NewInt(10),
// 				TotalTokenOut:    math.ZeroInt(),
// 			},
// 		},
// 	},
// }

// var inactiveLimitOrderTrancheList = []types.LimitOrderTranche{
// 	{
// 		PairID:           &types.PairID{Token0: "TokenA", Token1: "TokenB"},
// 		TokenIn:          "TokenB",
// 		TickIndex:        0,
// 		TrancheKey:       "0",
// 		TotalTokenIn:     math.NewInt(10),
// 		TotalTokenOut:    math.NewInt(10),
// 		ReservesTokenOut: math.NewInt(10),
// 		ReservesTokenIn:  math.NewInt(0),
// 	},
// 	{
// 		PairID:           &types.PairID{Token0: "TokenA", Token1: "TokenB"},
// 		TokenIn:          "TokenB",
// 		TickIndex:        0,
// 		TrancheKey:       "1",
// 		TotalTokenIn:     math.NewInt(10),
// 		TotalTokenOut:    math.NewInt(10),
// 		ReservesTokenOut: math.NewInt(10),
// 		ReservesTokenIn:  math.NewInt(0),
// 	},
// }

// var poolReservesList = []types.TickLiquidity{
// 	{
// 		Liquidity: &types.TickLiquidity_PoolReserves{
// 			PoolReserves: &types.PoolReserves{
// 				PairID: &types.PairID{
// 					Token0: "TokenA",
// 					Token1: "TokenB",
// 				},
// 				TokenIn:   "TokenB",
// 				TickIndex: 0,
// 				Fee:       1,
// 				Reserves:  math.NewInt(10),
// 			},
// 		},
// 	},
// 	{
// 		Liquidity: &types.TickLiquidity_PoolReserves{
// 			PoolReserves: &types.PoolReserves{
// 				PairID: &types.PairID{
// 					Token0: "TokenA",
// 					Token1: "TokenB",
// 				},
// 				TokenIn:   "TokenB",
// 				TickIndex: 0,
// 				Fee:       3,
// 				Reserves:  math.NewInt(10),
// 			},
// 		},
// 	},
// }

// var limitOrderTrancheUserList = []types.LimitOrderTrancheUser{
// 	{
// 		PairID:          &types.PairID{Token0: "TokenA", Token1: "TokenB"},
// 		Token:           "TokenA",
// 		TickIndex:       1,
// 		TrancheKey:      "0",
// 		Address:         testAddress.String(),
// 		SharesOwned:     math.NewInt(10),
// 		SharesWithdrawn: math.NewInt(0),
// 		SharesCancelled: math.NewInt(0),
// 	},
// 	{
// 		PairID:          &types.PairID{Token0: "TokenA", Token1: "TokenB"},
// 		Token:           "TokenB",
// 		TickIndex:       20,
// 		TrancheKey:      "1",
// 		Address:         testAddress.String(),
// 		SharesOwned:     math.NewInt(10),
// 		SharesWithdrawn: math.NewInt(0),
// 		SharesCancelled: math.NewInt(0),
// 	},
// }

// var genesisState types.GenesisState = types.GenesisState{
// 	TickLiquidityList:             append(poolReservesList, limitOrderTrancheList...),
// 	LimitOrderTrancheUserList:     limitOrderTrancheUserList,
// 	InactiveLimitOrderTrancheList: inactiveLimitOrderTrancheList,
// }

// func (s *QueryTestSuite) TestQueryCmdListTickLiquidity() {
// 	val := s.network.Validators[0]
// 	clientCtx := val.ClientCtx
// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expErr    bool
// 		expErrMsg string
// 		expOutput []types.TickLiquidity
// 	}{
// 		{
// 			name:      "valid",
// 			args:      []string{"TokenA<>TokenB", "TokenB"},
// 			expOutput: append(poolReservesList, limitOrderTrancheList...),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cmd := dexclient.CmdListTickLiquidity()
// 			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expErr {
// 				require.Error(s.T(), err)
// 				require.Contains(s.T(), out.String(), tc.expErrMsg)
// 			} else {
// 				require.NoError(s.T(), err)
// 				var res types.QueryAllTickLiquidityResponse
// 				require.NoError(s.T(), clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
// 				require.NotEmpty(s.T(), res)
// 				require.Equal(s.T(), tc.expOutput, res.TickLiquidity)
// 			}
// 		})
// 	}
// }

// func (s *QueryTestSuite) TestQueryCmdShowLimitOrderTranche() {
// 	val := s.network.Validators[0]
// 	clientCtx := val.ClientCtx
// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expErr    bool
// 		expErrMsg string
// 		expOutput types.LimitOrderTranche
// 	}{
// 		// show-limit-order-tranche [pair-id] [tick-index] [token-in] [tranche-key]
// 		{
// 			name:      "valid",
// 			args:      []string{"TokenA<>TokenB", "1", "TokenB", "0"},
// 			expOutput: *limitOrderTrancheList[0].GetLimitOrderTranche(),
// 		},
// 		{
// 			name:      "invalid pair",
// 			args:      []string{"TokenC<>TokenB", "20", "TokenB", "1"},
// 			expErr:    true,
// 			expErrMsg: "key not found",
// 		},
// 		{
// 			name:      "too many parameters",
// 			args:      []string{"TokenA<>B", "20", "TokenB", "1", "10"},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 4 arg(s), received 5",
// 		},
// 		{
// 			name:      "no parameters",
// 			args:      []string{},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 4 arg(s), received 0",
// 		},
// 		{
// 			name:      "too few parameters",
// 			args:      []string{"TokenA<>B", "20", "TokenB"},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 4 arg(s), received 3",
// 		},
// 	}
// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cmd := dexclient.CmdShowLimitOrderTranche()
// 			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expErr {
// 				require.Error(s.T(), err)
// 				require.Contains(s.T(), out.String(), tc.expErrMsg)
// 			} else {
// 				require.NoError(s.T(), err)
// 				var res types.QueryGetLimitOrderTrancheResponse
// 				require.NoError(s.T(), clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
// 				require.NotEmpty(s.T(), res)
// 				require.Equal(s.T(), tc.expOutput, res.LimitOrderTranche)
// 			}
// 		})
// 	}
// }

// func (s *QueryTestSuite) TestQueryCmdShowLimitOrderTrancheUser() {
// 	val := s.network.Validators[0]
// 	clientCtx := val.ClientCtx
// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expErr    bool
// 		expErrMsg string
// 		expOutput types.LimitOrderTrancheUser
// 	}{
// 		// "show-limit-order-tranche-user [address] [tranche-key]"
// 		{
// 			name:      "valid",
// 			args:      []string{testAddress.String(), "0"},
// 			expOutput: limitOrderTrancheUserList[0],
// 		},
// 		{
// 			name:      "invalid pair",
// 			args:      []string{testAddress.String(), "BADKEY"},
// 			expErr:    true,
// 			expErrMsg: "key not found",
// 		},
// 		{
// 			name:      "too many parameters",
// 			args:      []string{testAddress.String(), "0", "EXTRAARG"},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 2 arg(s), received 3",
// 		},
// 		{
// 			name:      "no parameters",
// 			args:      []string{},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 2 arg(s), received 0",
// 		},
// 		{
// 			name:      "too few parameters",
// 			args:      []string{testAddress.String()},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 2 arg(s), received 1",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cmd := dexclient.CmdShowLimitOrderTrancheUser()
// 			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expErr {
// 				require.Error(s.T(), err)
// 				require.Contains(s.T(), out.String(), tc.expErrMsg)
// 			} else {
// 				require.NoError(s.T(), err)

// 				var res types.QueryGetLimitOrderTrancheUserResponse
// 				require.NoError(s.T(), clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
// 				require.NotEmpty(s.T(), res)
// 				require.Equal(s.T(), tc.expOutput, res.LimitOrderTrancheUser)
// 			}
// 		})
// 	}
// }

// func (s *QueryTestSuite) TestQueryCmdListLimitOrderTrancheUser() {
// 	val := s.network.Validators[0]
// 	clientCtx := val.ClientCtx
// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expErr    bool
// 		expErrMsg string
// 		expOutput []types.LimitOrderTrancheUser
// 	}{
// 		{
// 			name:      "valid",
// 			args:      []string{},
// 			expOutput: limitOrderTrancheUserList,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cmd := dexclient.CmdListLimitOrderTrancheUser()
// 			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expErr {
// 				require.Error(s.T(), err)
// 				require.Contains(s.T(), out.String(), tc.expErrMsg)
// 			} else {
// 				require.NoError(s.T(), err)

// 				var res types.QueryAllLimitOrderTrancheUserResponse
// 				require.NoError(s.T(), clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
// 				require.NotEmpty(s.T(), res)
// 				require.Equal(s.T(), tc.expOutput, res.LimitOrderTrancheUser)
// 			}
// 		})
// 	}
// }

// func (s *QueryTestSuite) TestQueryCmdListInactiveLimitOrderTranche() {
// 	val := s.network.Validators[0]
// 	clientCtx := val.ClientCtx
// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expErr    bool
// 		expErrMsg string
// 		expOutput []types.LimitOrderTranche
// 	}{
// 		{
// 			name:      "valid",
// 			args:      []string{},
// 			expOutput: inactiveLimitOrderTrancheList,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cmd := dexclient.CmdListInactiveLimitOrderTranche()
// 			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expErr {
// 				require.Error(s.T(), err)
// 				require.Contains(s.T(), out.String(), tc.expErrMsg)
// 			} else {
// 				require.NoError(s.T(), err)

// 				var res types.QueryAllInactiveLimitOrderTrancheResponse
// 				require.NoError(s.T(), clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
// 				require.NotEmpty(s.T(), res)
// 				require.Equal(s.T(), tc.expOutput, res.InactiveLimitOrderTranche)
// 			}
// 		})
// 	}
// }

// func (s *QueryTestSuite) TestQueryCmdShowInactiveLimitOrderTranche() {
// 	val := s.network.Validators[0]
// 	clientCtx := val.ClientCtx
// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expErr    bool
// 		expErrMsg string
// 		expOutput types.LimitOrderTranche
// 	}{
// 		// show-filled-limit-order-tranche [pair-id] [token-in] [tick-index] [tranche-index]",
// 		{
// 			name:      "valid",
// 			args:      []string{"TokenA<>TokenB", "TokenB", "0", "0"},
// 			expOutput: inactiveLimitOrderTrancheList[0],
// 		},
// 		{
// 			name:      "invalid pair",
// 			args:      []string{"TokenC<>TokenB", "TokenB", "0", "0"},
// 			expErr:    true,
// 			expErrMsg: "key not found",
// 		},
// 		{
// 			name:      "too many parameters",
// 			args:      []string{"TokenC<>TokenB", "TokenB", "0", "0", "Extra arg"},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 4 arg(s), received 5",
// 		},
// 		{
// 			name:      "no parameters",
// 			args:      []string{},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 4 arg(s), received 0",
// 		},
// 		{
// 			name:      "too few parameters",
// 			args:      []string{"TokenC<>TokenB", "TokenB", "0"},
// 			expErr:    true,
// 			expErrMsg: "Error: accepts 4 arg(s), received 3",
// 		},
// 	}
// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cmd := dexclient.CmdShowInactiveLimitOrderTranche()
// 			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expErr {
// 				require.Error(s.T(), err)
// 				require.Contains(s.T(), out.String(), tc.expErrMsg)
// 			} else {
// 				require.NoError(s.T(), err)
// 				var res types.QueryGetInactiveLimitOrderTrancheResponse
// 				require.NoError(s.T(), clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
// 				require.NotEmpty(s.T(), res)
// 				require.Equal(s.T(), tc.expOutput, res.InactiveLimitOrderTranche)
// 			}
// 		})
// 	}
// }
