package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

// CallSudoForSubscriptionType calls sudo for all contracts subscribed to given `hookType`.
// Returns error in cases where marshalling error occurred (should never happen, since we control it) or
// when any error in contract happened.
// Important that because some calls are coming from BeginBlocker/EndBlocker, any errors in contracts can halt the chain.
func (k Keeper) CallSudoForSubscriptionType(ctx context.Context, hookType types.HookType, msg any) error {
	if err := k.DoCallSudoForSubscriptionType(ctx, hookType, msg); err != nil {
		return errors.Wrapf(err, "failed to call sudo for subscriptions for hookType=%s", hookType)
	}

	return nil
}

func (k Keeper) DoCallSudoForSubscriptionType(ctx context.Context, hookType types.HookType, msg any) error {
	contractAddresses := k.GetSubscribedAddressesForHookType(ctx, hookType)

	if len(contractAddresses) == 0 {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	msgJsonBz, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal sudo subscription msg: %v", err)
	}

	for _, contractAddress := range contractAddresses {
		// As we're using ctx here (no cached context!), any errors will be returned from hooks as it is.
		// This means it can potentially halt the chain OR abort the transactions depending on where hook was called from.
		accContractAddress, err := sdk.AccAddressFromBech32(contractAddress)
		if err != nil {
			return errors.Wrapf(err, "could not parse acc address from bech32 for harpoon sudo call")
		}
		_, err = k.wasmKeeper.Sudo(sdkCtx, accContractAddress, msgJsonBz)
		if err != nil {
			sdkCtx.Logger().Error("execute harpoon subscription hook error: failed to execute contract msg",
				"contract_address", contractAddress,
				"error", err,
			)
			return err
		}
	}

	return nil
}
