package wasmbinding

import sdk "github.com/cosmos/cosmos-sdk/types"

// GetInterchainQueryResult is a function, not method, so the message_plugin can use it
func (qp *QueryPlugin) GetInterchainQueryResult(ctx sdk.Context, queryID uint64) (string, error) {
	// TODO
	return "", nil
}

func (qp *QueryPlugin) GetInterchainAccountAddress(ctx sdk.Context, ownerAddress, connectionID string) (string, error) {
	//TODO
	return "", nil
}
