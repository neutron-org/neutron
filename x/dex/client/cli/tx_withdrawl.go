package cli

import (
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

func CmdWithdrawal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdrawal [receiver] [token-a] [token-b] [list of shares-to-remove] [list of tick-index] [list of fees] ",
		Short:   "Broadcast message withdrawal",
		Example: "withdrawal alice tokenA tokenB 100,50 [-10,5] 1,1 --from alice",
		Args:    cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argReceiver := args[0]
			argTokenA := args[1]
			argTokenB := args[2]
			argSharesToRemove := strings.Split(args[3], ",")

			if strings.HasPrefix(args[4], "[") && strings.HasSuffix(args[4], "]") {
				args[4] = strings.TrimPrefix(args[4], "[")
				args[4] = strings.TrimSuffix(args[4], "]")
			}
			argTickIndexes := strings.Split(args[4], ",")
			argFees := strings.Split(args[5], ",")

			var SharesToRemoveInt []math.Int
			var TicksIndexesInt []int64
			var FeesUint []uint64
			for _, s := range argSharesToRemove {
				sharesToRemoveInt, ok := math.NewIntFromString(s)

				if !ok {
					return sdkerrors.Wrapf(types.ErrIntOverflowTx, "Integer Overflow for shares-to-remove")
				}

				SharesToRemoveInt = append(SharesToRemoveInt, sharesToRemoveInt)
			}

			for _, s := range argTickIndexes {
				TickIndexInt, err := strconv.ParseInt(s, 10, 0)
				if err != nil {
					return err
				}

				TicksIndexesInt = append(TicksIndexesInt, TickIndexInt)
			}

			for _, s := range argFees {
				feeInt, err := strconv.ParseUint(s, 10, 0)
				if err != nil {
					return err
				}

				FeesUint = append(FeesUint, feeInt)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawal(
				clientCtx.GetFromAddress().String(),
				argReceiver,
				argTokenA,
				argTokenB,
				SharesToRemoveInt,
				TicksIndexesInt,
				FeesUint,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
