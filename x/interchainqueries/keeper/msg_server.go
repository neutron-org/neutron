package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) RegisterInterchainQuery(goCtx context.Context, msg *types.MsgRegisterInterchainQuery) (*types.MsgRegisterInterchainQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lastID := k.GetLastRegisteredQueryKey(ctx)
	lastID += 1

	registeredQuery := types.RegisteredQuery{
		Id:               lastID,
		QueryData:        msg.QueryData,
		QueryType:        msg.QueryType,
		ZoneId:           msg.ZoneId,
		UpdatePeriod:     msg.UpdatePeriod,
		LastLocalHeight:  0,
		LastRemoteHeight: 0,
	}

	k.SaveQuery(ctx, registeredQuery)

	return &types.MsgRegisterInterchainQueryResponse{Id: lastID}, nil
}

func (k msgServer) SubmitQueryResult(goCtx context.Context, msg *types.MsgSubmitQueryResult) (*types.MsgSubmitQueryResultResponse, error) {
	for _, result := range msg.KVResults {
		proof, err := types2.ConvertProofs(result.Proof)
		if err != nil {
			return nil, fmt.Errorf("failed to convert crypto.ProofOps to MerkleProof: %w", err)
		}

		if err := proof.VerifyMembership(types2.GetSDKSpecs(), nil, nil, nil); err != nil {
			return nil, fmt.Errorf("failed to verify proof: %w", err)
		}
	}
	return nil, nil
}

var _ types.MsgServer = msgServer{}
