package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	icqtypes "github.com/neutron-org/neutron/v8/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v8/x/state-verifier/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group state-verifier queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryConsensusState())
	cmd.AddCommand(CmdVefifyStorageValues())

	return cmd
}

func CmdQueryConsensusState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consensus-state [height]",
		Short:   "returns saved consensus state by a particular height",
		Example: "consensus-state 12345",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			request := &types.QueryConsensusStateRequest{
				Height: height,
			}

			res, err := queryClient.QueryConsensusState(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdVefifyStorageValues() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-storage-values [height] [storage-values-json-file]",
		Short: "verifies storage values that saved in a file for a particular height",
		Example: `neutrond verify-storage-values 12345 values.json

Example format of the values.json file:
"
[
  {
	"storage_prefix": "acc",
	"key": "AdrcG6pcvJ42SF5pCqQz+jM5bpuH",
	"value": null,
	"Proof": {
	  "ops": [
		{
		  "type": "ics23:iavl",
		  "key": "AdrcG6pcvJ42SF5pCqQz+jM5bpuH",
		  "data": "EqMHChUB2twbqly8njZIXmkKpDP6Mzlum4cS1AMKIQHZoqcVc0mY8VJPe/1H3OzI/ATEwZPK6PBMgONwaOmtzBJqCiAvY29zbW9zLmF1dGgudjFiZXRhMS5CYXNlQWNjb3VudBJGCkJuZXV0cm9uMW14MzJ3OXRuZnh2MHo1ajAwMDc1MGg4dmVyN3FmM3hwajA5dzN1enZzcjNocTY4ZjRoeHF3ZzA2dGEYGBoLCAEYASABKgMAAgIiKwgBEgQCBAIgGiEgR9cz8KTJwg1r42MLHThCK8MzirizXBPfocsHWqc+hVYiKwgBEgQECAIgGiEgCJWoCKSWlqFHiqAKn5TbrPosv3hZ+jh9OyRfLvVVfJYiKQgBEiUGEAIgz8Snw8OaCo1jA3kj06N8pswAJBzEI4yPK8/VS0g+SdogIisIARIECCACIBohIIci55eyWy6DOHZv6gl5Vz4jaWNFPOs/ufVD/I27Q7wTIisIARIECkACIBohIEW4dTl4yQvJr+nI0Y/VuKgmYjAQFCOn66vOBHXkaPqtIisIARIEDGQCIBohIJu1qyrg2QndMQQol/GeRVnAzhf3RsMrWP8A5/44x9Z4IioIARImDqQBAiAWbdtCGTVWVFf7sIfjrzXa57hyHvDbg+yAP8isbFpfwyAasgMKFQHcreKmUkBD/tqOAeJ3PI0p1IFYiRJWCiAvY29zbW9zLmF1dGgudjFiZXRhMS5CYXNlQWNjb3VudBIyCi5uZXV0cm9uMW1qazc5ZmpqZ3BwbGFrNXdxODM4dzB5ZDk4Mmd6a3lmOGZ4dTh1GAUaCwgBGAEgASoDAAICIikIARIlAgQCIF2MLUlhpabEuyhYrS/hgNoXlrRhge/lXYU2xsRYKLstICIrCAESBAQIAiAaISAIlagIpJaWoUeKoAqflNus+iy/eFn6OH07JF8u9VV8liIpCAESJQYQAiDPxKfDw5oKjWMDeSPTo3ymzAAkHMQjjI8rz9VLSD5J2iAiKwgBEgQIIAIgGiEghyLnl7JbLoM4dm/qCXlXPiNpY0U86z+59UP8jbtDvBMiKwgBEgQKQAIgGiEgRbh1OXjJC8mv6cjRj9W4qCZiMBAUI6frq84EdeRo+q0iKwgBEgQMZAIgGiEgm7WrKuDZCd0xBCiX8Z5FWcDOF/dGwytY/wDn/jjH1ngiKggBEiYOpAECIBZt20IZNVZUV/uwh+OvNdrnuHIe8NuD7IA/yKxsWl/DIA=="
		},
		{
		  "type": "ics23:simple",
		  "key": "YWNj",
		  "data": "CqgCCgNhY2MSIGqWHw3wwQKzzfbeomw/rZSu3SB5JUfxBlw/TjXRbxnOGgkIARgBIAEqAQAiJwgBEgEBGiAZo7ews4BQuDJEY3t1nLp5AGZtIs9kJ5owO72sXm/l5yInCAESAQEaILdSKvzUxsoElid97tdX+QUoRGzdVxvwjX2xfLR1zSxzIicIARIBARogROr0OpkBQynhUpSp4L756nD/osZ7tN54z0RdSMzfxB4iJwgBEgEBGiAmWWRA2IkD0cDt8VW2hyyrym2I6v9Gmg1g6ORp6UTOVSInCAESAQEaICZr8cBZ1NbH1gwbehyDs418wgPLgSUnfpQRFQVDgpZvIicIARIBARogN0IwDGF6wDLX/b73BXuhYru5z599l7C70cCSbsud+8k="
		}
	  ]
	}
  }
]
"
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			storageValuesFile, err := os.ReadFile(args[1])
			if err != nil {
				return errors.Wrap(err, "failed to read storage values file")
			}

			var storageValues []*icqtypes.StorageValue
			if err := json.Unmarshal(storageValuesFile, &storageValues); err != nil {
				return errors.Wrap(err, "failed to unmarshal storage values file")
			}

			request := &types.QueryVerifyStateValuesRequest{
				Height:        height,
				StorageValues: storageValues,
			}

			res, err := queryClient.VerifyStateValues(context.Background(), request)
			if err != nil {
				return errors.Wrapf(err, "failed to verify storage values for height %d", height)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
