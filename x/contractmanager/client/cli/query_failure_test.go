package cli_test

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"

	appparams "github.com/neutron-org/neutron/v6/app/params"

	"github.com/neutron-org/neutron/v6/app/config"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/testutil/contractmanager/network"
	"github.com/neutron-org/neutron/v6/x/contractmanager/client/cli"
	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

func networkWithFailureObjects(t *testing.T, n int) (*network.Network, []types.Failure) {
	sdktypes.DefaultBondDenom = appparams.DefaultDenom

	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < n; i++ {
		_, err := rand.Read(pub.Key)
		require.NoError(t, err)
		acc := sdktypes.AccAddress(pub.Address())
		failure := types.Failure{
			Address:     acc.String(),
			SudoPayload: []byte("&channeltypes.Packet{}"),
		}
		nullify.Fill(&failure)
		state.FailuresList = append(state.FailuresList, failure)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.FailuresList
}

func TestAddressFailures(t *testing.T) {
	_ = config.GetDefaultConfig()

	net, objs := networkWithFailureObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc    string
		idIndex string

		args []string
		err  error
		obj  []types.Failure
	}{
		{
			desc:    "found",
			idIndex: objs[0].Address,

			args: common,
			obj:  []types.Failure{objs[0]},
		},
		{
			desc:    "not found",
			idIndex: strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idIndex,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdFailures(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryFailuresResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Failures)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Failures),
				)
			}
		})
	}
}

func TestListFailure(t *testing.T) {
	net, objs := networkWithFailureObjects(t, 5)

	ctx := net.Validators[0].ClientCtx
	request := func(next []byte, offset, limit uint64, total bool) []string {
		args := []string{
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		}
		if next == nil {
			args = append(args, fmt.Sprintf("--%s=%d", flags.FlagOffset, offset))
		} else {
			args = append(args, fmt.Sprintf("--%s=%s", flags.FlagPageKey, next))
		}
		args = append(args, fmt.Sprintf("--%s=%d", flags.FlagLimit, limit))
		if total {
			args = append(args, fmt.Sprintf("--%s", flags.FlagCountTotal))
		}
		return args
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(objs); i += step {
			args := request(nil, uint64(i), uint64(step), false) //nolint:gosec
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdFailures(), args)
			require.NoError(t, err)
			var resp types.QueryFailuresResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.Failures), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.Failures),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(objs); i += step {
			args := request(next, 0, uint64(step), false) //nolint:gosec
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdFailures(), args)
			require.NoError(t, err)
			var resp types.QueryFailuresResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.Failures), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.Failures),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		args := request(nil, 0, uint64(len(objs)), true)
		out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdFailures(), args)
		require.NoError(t, err)
		var resp types.QueryFailuresResponse
		require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		require.NoError(t, err)
		require.Equal(t, uint64(len(objs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			nullify.Fill(objs),
			nullify.Fill(resp.Failures),
		)
	})
}
