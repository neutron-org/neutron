package cli_test

import (
	"testing"

	"github.com/neutron-org/neutron/app"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/testutil/contractmanager/network"
	"github.com/neutron-org/neutron/x/feerefunder/client/cli"
	"github.com/neutron-org/neutron/x/feerefunder/types"
)

func feeRefunderNetwork(t *testing.T, feeInfo types.Fee) *network.Network {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.DefaultGenesis()
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], state))
	state.FeeInfos = []types.FeeInfo{{
		Payer: "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh",
		PacketId: types.PacketID{
			ChannelId: "channel-0",
			PortId:    "transfer",
			Sequence:  111,
		},
		Fee: feeInfo,
	}}
	buf, err := cfg.Codec.MarshalJSON(state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg)
}

func TestQueryFeeInfo(t *testing.T) {
	_ = app.GetDefaultConfig()

	feeInfo := types.Fee{
		RecvFee:    sdk.NewCoins(),
		AckFee:     sdk.NewCoins(sdk.NewCoin("untrn", sdk.NewInt(1001))),
		TimeoutFee: sdk.NewCoins(sdk.NewCoin("untrn", sdk.NewInt(2001))),
	}
	net := feeRefunderNetwork(t, feeInfo)

	ctx := net.Validators[0].ClientCtx
	validArgs := []string{
		"transfer",
		"channel-0",
		"111",
	}
	out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdFeeInfo(), validArgs)
	require.NoError(t, err)
	var resp types.FeeInfoResponse
	require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
	require.Equal(t, feeInfo, resp.GetFeeInfo().GetFee())

	invalidArgs := []string{
		"wrongport",
		"channel-0",
		"111",
	}
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdFeeInfo(), invalidArgs)
	require.ErrorContains(t, err, "no fee info found for port_id = wrongport, channel_id=channel-0, sequence=111")
}
