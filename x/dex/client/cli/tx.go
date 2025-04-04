package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

var DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdDeposit())
	cmd.AddCommand(CmdWithdrawal())
	cmd.AddCommand(CmdPlaceLimitOrder())
	cmd.AddCommand(CmdWithdrawFilledLimitOrder())
	cmd.AddCommand(CmdCancelLimitOrder())
	cmd.AddCommand(CmdMultiHopSwap())
	// this line is used by starport scaffolding # 1

	return cmd
}
