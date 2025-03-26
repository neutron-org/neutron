package cli

import (
	"log"
	"strconv"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		//nolint:lll
		Use:     "deposit [receiver] [token-a] [token-b] [list of amount-0] [list of amount-1] [list of tick-index] [list of fees] [disable_autoswap], [fail_tx_on_BEL]",
		Short:   "Broadcast message deposit",
		Example: "deposit alice tokenA tokenB 100,0 0,50 [-10,5] 1,1 false,false false,false --from alice",
		Args:    cobra.ExactArgs(9),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argReceiver := args[0]
			argTokenA := args[1]
			argTokenB := args[2]
			argAmountsA := strings.Split(args[3], ",")
			argAmountsB := strings.Split(args[4], ",")

			if args[5] == "-" {
				log.Printf("\"this is a test\": %v\n", "this is a test")
			}

			if strings.HasPrefix(args[5], "[") && strings.HasSuffix(args[5], "]") {
				args[5] = strings.TrimPrefix(args[5], "[")
				args[5] = strings.TrimSuffix(args[5], "]")
			}
			argTicksIndexes := strings.Split(args[5], ",")

			argFees := strings.Split(args[6], ",")
			argAutoswapOptions := strings.Split(args[7], ",")
			argFailTxOptions := strings.Split(args[8], ",")

			swapOnDeposit, err := cmd.Flags().GetBool(FlagSwapOnDeposit)
			if err != nil {
				return err
			}

			var AmountsA []math.Int
			var AmountsB []math.Int
			var TicksIndexesInt []int64
			var FeesUint []uint64
			var DepositOptions []*types.DepositOptions

			for _, s := range argAmountsA {
				amountA, ok := math.NewIntFromString(s)
				if !ok {
					return sdkerrors.Wrapf(types.ErrIntOverflowTx, "Integer overflow for amountsA")
				}

				AmountsA = append(AmountsA, amountA)
			}

			for _, s := range argAmountsB {
				amountB, ok := math.NewIntFromString(s)
				if !ok {
					return sdkerrors.Wrapf(types.ErrIntOverflowTx, "Integer overflow for amountsB")
				}

				AmountsB = append(AmountsB, amountB)
			}

			for _, s := range argTicksIndexes {
				TickIndexInt, err := strconv.ParseInt(s, 10, 0)
				if err != nil {
					return err
				}

				TicksIndexesInt = append(TicksIndexesInt, TickIndexInt)
			}

			for _, s := range argFees {
				FeeInt, err := strconv.ParseUint(s, 10, 0)
				if err != nil {
					return err
				}

				FeesUint = append(FeesUint, FeeInt)
			}

			for i, s := range argAutoswapOptions {
				disableAutoswap, err := strconv.ParseBool(s)
				if err != nil {
					return err
				}
				failTx, err := strconv.ParseBool(argFailTxOptions[i])
				if err != nil {
					return err
				}

				DepositOptions = append(DepositOptions, &types.DepositOptions{
					DisableAutoswap: disableAutoswap,
					FailTxOnBel:     failTx,
					SwapOnDeposit:   swapOnDeposit,
				})
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeposit(
				clientCtx.GetFromAddress().String(),
				argReceiver,
				argTokenA,
				argTokenB,
				AmountsA,
				AmountsB,
				TicksIndexesInt,
				FeesUint,
				DepositOptions,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetSwapOnDeposit())

	return cmd
}
