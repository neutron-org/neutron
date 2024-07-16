package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdPlaceLimitOrder())
	cmd.AddCommand(CmdWithdrawFilledLimitOrder())
	cmd.AddCommand(CmdMultiHopSwap())
	// this line is used by starport scaffolding # 1

	return cmd
}
