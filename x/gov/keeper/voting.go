package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/neutron-org/neutron/x/gov/types"
)

// queryVotingPowers queries the dao contract of user voting powers based on the given query msg
func queryVotingPowers(ctx sdk.Context, k wasmtypes.ViewKeeper, contractAddr sdk.AccAddress) (types.VotingPowersResponse, error) {
	var votingPowersResponse types.VotingPowersResponse

	req, err := json.Marshal(&types.QueryMsg{VotingPowers: &types.VotingPowersQuery{}})
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrFailedToQueryVesting, "failed to marshal query request: %s", err)
	}

	res, err := k.QuerySmart(ctx, contractAddr, req)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrFailedToQueryVesting, "query returned error: %s", err)
	}

	err = json.Unmarshal(res, &votingPowersResponse)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrFailedToQueryVesting, "failed to unmarshal query response: %s", err)
	}

	return votingPowersResponse, nil
}

// incrementVotingPowers increments the voting power counter based on the contract query response
//
// NOTE: This function modifies the `tokensLocked` and `totalTokensAmount` variables in place.
func incrementVotingPowers(votingPowersResponse types.VotingPowersResponse, tokensLocked map[string]sdk.Int, totalTockensLocked *sdk.Int) error {
	for _, item := range votingPowersResponse {
		if _, ok := tokensLocked[item.User]; ok {
			return sdkerrors.Wrapf(types.ErrFailedToQueryVesting, "query response contains duplicate address: %s", item.User)
		}

		tokensLocked[item.User] = sdk.Int(item.VotingPower)
		*totalTockensLocked = totalTockensLocked.Add(sdk.Int(item.VotingPower))
	}

	return nil
}

// GetTokensInDao queries the vesting contract for an array of users who have tokens locked in the
// contract and their respective amount, as well as computing the total amount of locked tokens.
func GetTokensInDao(ctx sdk.Context, k wasmtypes.ViewKeeper, contractAddr sdk.AccAddress) (map[string]sdk.Int, sdk.Int, error) {
	tokensLocked := make(map[string]sdk.Int)
	totalTokenAmount := sdk.ZeroInt()

	votingPowersResponse, err := queryVotingPowers(ctx, k, contractAddr)
	if err != nil {
		return nil, sdk.ZeroInt(), err
	}

	if err = incrementVotingPowers(votingPowersResponse, tokensLocked, &totalTokenAmount); err != nil {
		return nil, sdk.ZeroInt(), err
	}

	return tokensLocked, totalTokenAmount, nil
}

// MustGetTokensInVesting is the same with `GetTokensInDao`, but panics on error
func MustGetTokensInVesting(ctx sdk.Context, k wasmtypes.ViewKeeper, contractAddr sdk.AccAddress) (map[string]sdk.Int, sdk.Int) {
	tokensLocked, totalTokensLocked, err := GetTokensInDao(ctx, k, contractAddr)
	if err != nil {
		panic(fmt.Sprintf("failed to tally vote: %s", err))
	}

	return tokensLocked, totalTokensLocked
}
