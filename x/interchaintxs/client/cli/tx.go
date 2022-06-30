package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
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

	cmd.AddCommand(RegisterInterchainAccountCmd())
	cmd.AddCommand(SubmitTxCmd())

	return cmd
}

func RegisterInterchainAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-interchain-account [connection-id] [owner]",
		Short:   "Register an interchain account",
		Aliases: []string{"register", "r"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fromAddress := clientCtx.GetFromAddress()
			connectionID := args[0]
			owner := args[1]

			msg := types.MsgRegisterInterchainAccount{
				FromAddress:  fromAddress.String(),
				ConnectionId: connectionID,
				Owner:        owner,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func SubmitTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "submit-tx [connection-id] [owner] [path/to/sdk_msgs.json]",
		Short:   "Submit interchain tx",
		Aliases: []string{"submit", "s"},
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress()
			connectionID := args[0]
			owner := args[1]
			pathToMsgs := args[2]

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var txMsgs []sdk.Msg
			if err := cdc.UnmarshalInterfaceJSON([]byte(pathToMsgs), &txMsgs); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(pathToMsgs)
				if err != nil {
					return fmt.Errorf("json input was not provided; failed to read file with tx messages: %w", err)
				}

				var rawTxMsgs struct {
					Msgs []json.RawMessage `json:"msgs"`
				}

				if err := json.Unmarshal(contents, &rawTxMsgs); err != nil {
					return fmt.Errorf("cannot unmarshal msgs array: %w", err)
				}

				for _, txMsg := range rawTxMsgs.Msgs {
					var sdkMsg sdk.Msg
					if err := cdc.UnmarshalInterfaceJSON(txMsg, &sdkMsg); err != nil {
						return fmt.Errorf("cannot unmarshal submessage: %w", err)
					}
					txMsgs = append(txMsgs, sdkMsg)
				}
			}

			var anyMsgs []*codectypes.Any
			for idx, msg := range txMsgs {
				anyMsg, err := types.PackTxMsgAny(msg)
				if err != nil {
					return fmt.Errorf("failed to PackTxMsgAny msg #%d: %s", idx, err)
				}
				anyMsgs = append(anyMsgs, anyMsg)
			}

			msg := types.MsgSubmitTx{
				FromAddress:  sender.String(),
				ConnectionId: connectionID,
				Owner:        owner,
				Msgs:         anyMsgs,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
