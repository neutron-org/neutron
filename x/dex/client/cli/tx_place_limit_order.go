package cli

import (
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/spf13/cobra"
)

func CmdPlaceLimitOrder() *cobra.Command {
	cmd := &cobra.Command{
		//nolint:lll
		Use:     "place-limit-order [receiver] [token-in] [token-out] [tick-index] [amount-in] ?[order-type] ?[expirationTime] ?(--max-amout-out)",
		Short:   "Broadcast message PlaceLimitOrder",
		Example: "place-limit-order alice tokenA tokenB [-10] tokenA 50 GOOD_TIL_TIME '01/02/2006 15:04:05' --max-amount-out 20 --from alice",
		Args:    cobra.RangeArgs(5, 7),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argReceiver := args[0]
			argTokenIn := args[1]
			argTokenOut := args[2]
			if strings.HasPrefix(args[3], "[") && strings.HasSuffix(args[3], "]") {
				args[3] = strings.TrimPrefix(args[3], "[")
				args[3] = strings.TrimSuffix(args[3], "]")
			}
			argTickIndex := args[3]
			argTickIndexInt, err := strconv.ParseInt(argTickIndex, 10, 0)
			if err != nil {
				return err
			}
			argAmountIn := args[4]

			amountInInt, ok := math.NewIntFromString(argAmountIn)
			if !ok {
				return sdkerrors.Wrapf(types.ErrIntOverflowTx, "Integer overflow for amount-in")
			}

			orderType := types.LimitOrderType_GOOD_TIL_CANCELLED
			if len(args) >= 6 {
				orderTypeInt, ok := types.LimitOrderType_value[args[5]]
				if !ok {
					return types.ErrInvalidOrderType
				}
				orderType = types.LimitOrderType(orderTypeInt)
			}

			var goodTil *time.Time
			if len(args) == 7 {
				const timeFormat = "01/02/2006 15:04:05"
				tm, err := time.Parse(timeFormat, args[6])
				if err != nil {
					return sdkerrors.Wrapf(types.ErrInvalidTimeString, err.Error())
				}
				goodTil = &tm
			}

			maxAmountOutArg, err := cmd.Flags().GetString(FlagMaxAmountOut)
			if err != nil {
				return err
			}
			var maxAmountOutIntP *math.Int = nil
			if maxAmountOutArg != "" {
				maxAmountOutInt, ok := math.NewIntFromString(maxAmountOutArg)
				if !ok {
					return sdkerrors.Wrapf(
						types.ErrIntOverflowTx,
						"Integer overflow for max-amount-out",
					)
				}
				maxAmountOutIntP = &maxAmountOutInt
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgPlaceLimitOrder(
				clientCtx.GetFromAddress().String(),
				argReceiver,
				argTokenIn,
				argTokenOut,
				argTickIndexInt,
				amountInInt,
				orderType,
				goodTil,
				maxAmountOutIntP,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetMaxAmountOut())

	return cmd
}
