package cli

import (
	
    "github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

func CmdCreateFailure() *cobra.Command {
    cmd := &cobra.Command{
		Use:   "create-failure [index] [contract-address] [ack-id] [ack-type]",
		Short: "Create a new failure",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
            // Get indexes
         indexIndex := args[0]
        
            // Get value arguments
		 argContractAddress := args[1]
		 argAckId := args[2]
		 argAckType := args[3]
		
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateFailure(
			    clientCtx.GetFromAddress().String(),
			    indexIndex,
                argContractAddress,
			    argAckId,
			    argAckType,
			    )
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

    return cmd
}

func CmdUpdateFailure() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-failure [index] [contract-address] [ack-id] [ack-type]",
		Short: "Update a failure",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
            // Get indexes
         indexIndex := args[0]
        
            // Get value arguments
		 argContractAddress := args[1]
		 argAckId := args[2]
		 argAckType := args[3]
		
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateFailure(
			    clientCtx.GetFromAddress().String(),
			    indexIndex,
                argContractAddress,
                argAckId,
                argAckType,
                )
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

    return cmd
}

func CmdDeleteFailure() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-failure [index]",
		Short: "Delete a failure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
             indexIndex := args[0]
            
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteFailure(
			    clientCtx.GetFromAddress().String(),
			    indexIndex,
                )
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

    return cmd
}