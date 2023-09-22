package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/neutron-org/neutron/utils/dcli"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := dcli.TxIndexCmd(types.ModuleName)

	dcli.AddTxCmd(cmd, NewCreateGaugeCmd)
	dcli.AddTxCmd(cmd, NewAddToGaugeCmd)
	dcli.AddTxCmd(cmd, NewStakeCmd)
	dcli.AddTxCmd(cmd, NewUnstakeCmd)

	return cmd
}

func CreateGaugeCmdBuilder(
	clientCtx client.Context,
	args []string,
	flags *pflag.FlagSet,
) (sdk.Msg, error) {
	// "create-gauge [pairTokenA] [pairTokenB] [startTick] [endTick] [coins] [numEpochs] [pricingTick]"
	pairID, err := dextypes.NewPairIDFromUnsorted(args[0], args[1])
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	startTick, err := dcli.ParseIntMaybeNegative(args[2], "startTick")
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	endTick, err := dcli.ParseIntMaybeNegative(args[3], "endTick")
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	coins, err := sdk.ParseCoinsNormalized(args[4])
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	var startTime time.Time
	timeStr, err := flags.GetString(FlagStartTime)
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}
	if timeStr == "" { // empty start time
		startTime = time.Unix(0, 0).UTC()
	} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
		startTime = time.Unix(timeUnix, 0).UTC()
	} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
		startTime = timeRFC
	} else { // invalid input
		return &types.MsgCreateGauge{}, errors.New("invalid start time format")
	}

	epochs, err := dcli.ParseUint(args[5], "numEpochs")
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	perpetual, err := flags.GetBool(FlagPerpetual)
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	if perpetual {
		epochs = 1
	}

	pricingTick, err := dcli.ParseIntMaybeNegative(args[6], "pricingTick")
	if err != nil {
		return &types.MsgCreateGauge{}, err
	}

	distributeTo := types.QueryCondition{
		PairID:    pairID,
		StartTick: startTick,
		EndTick:   endTick,
	}

	msg := types.NewMsgCreateGauge(
		epochs == 1,
		clientCtx.GetFromAddress(),
		distributeTo,
		coins,
		startTime,
		epochs,
		pricingTick,
	)

	return msg, nil
}

func NewCreateGaugeCmd() (*dcli.TxCliDesc, *types.MsgCreateGauge) {
	return &dcli.TxCliDesc{
		ParseAndBuildMsg: CreateGaugeCmdBuilder,
		Use:              "create-gauge [pairTokenA] [pairTokenB] [startTick] [endTick] [coins] [numEpochs] [pricingTick]",
		Short:            "create a gauge to distribute rewards to users",
		Long:             `{{.Short}}{{.ExampleHeader}} create-gauge TokenA TokenB [-10] 200 100TokenA,200TokenB 6 0 --start-time 2006-01-02T15:04:05Z07:00 --perpetual true`,
		Flags:            dcli.FlagDesc{OptionalFlags: []*pflag.FlagSet{FlagSetCreateGauge()}},
		NumArgs:          7,
	}, &types.MsgCreateGauge{}
}

func NewAddToGaugeCmd() (*dcli.TxCliDesc, *types.MsgAddToGauge) {
	return &dcli.TxCliDesc{
		Use:   "add-to-gauge [gauge_id] [coins]",
		Short: "add coins to gauge to distribute more rewards to users",
		Long:  `{{.Short}}{{.ExampleHeader}} add-to-gauge 1 TokenA,TokenB`,
	}, &types.MsgAddToGauge{}
}

func NewStakeCmd() (*dcli.TxCliDesc, *types.MsgStake) {
	return &dcli.TxCliDesc{
		Use:   "stake-tokens [coins]",
		Short: "stake tokens into stake pool from user account",
	}, &types.MsgStake{}
}

func UnstakeCmdBuilder(clientCtx client.Context, args []string, _ *pflag.FlagSet) (sdk.Msg, error) {
	// "unstake-tokens [poolID]:[coins] [poolID]:[coins] ..."
	unstakes := make([]*types.MsgUnstake_UnstakeDescriptor, 0, len(args))
	for i, unstake := range args {
		if strings.HasPrefix(unstake, "-") {
			// no more unstakes left, only flags
			break
		}

		parts := strings.Split(unstake, ":")
		if len(parts) != 2 {
			return &types.MsgUnstake{}, errors.New("invalid syntax for unstake tokens")
		}
		poolID, err := dcli.ParseUint(parts[0], fmt.Sprintf("poolID[%d]", i))
		if err != nil {
			return &types.MsgUnstake{}, err
		}

		coins, err := dcli.ParseCoins(parts[1], fmt.Sprintf("coins[%d]", i))
		if err != nil {
			return &types.MsgUnstake{}, err
		}

		unstakes = append(unstakes, types.NewMsgUnstakeDescriptor(poolID, coins))
	}

	return types.NewMsgUnstake(clientCtx.GetFromAddress(), unstakes), nil
}

func NewUnstakeCmd() (*dcli.TxCliDesc, *types.MsgUnstake) {
	return &dcli.TxCliDesc{
		Use:              "unstake-tokens [poolID]:[coins] [poolID]:[coins] ...",
		Short:            "Unstake tokens",
		ParseAndBuildMsg: UnstakeCmdBuilder,
		Long:             `{{.Short}}{{.ExampleHeader}} unstake-tokens 1:100TokenA 2:10TokenZ,20TokenB`,
	}, &types.MsgUnstake{}
}
