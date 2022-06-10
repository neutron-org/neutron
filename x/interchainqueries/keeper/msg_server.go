package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	tmtypes "github.com/tendermint/tendermint/types"
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
		Id:                lastID,
		QueryData:         msg.QueryData,
		QueryType:         msg.QueryType,
		ZoneId:            msg.ZoneId,
		UpdatePeriod:      msg.UpdatePeriod,
		ConnectionId:      msg.ConnectionId,
		LastEmittedHeight: uint64(ctx.BlockHeight()),
	}

	k.SetLastRegisteredQueryKey(ctx, lastID)
	if err := k.SaveQuery(ctx, registeredQuery); err != nil {
		return nil, fmt.Errorf("failed to save query: %w", err)
	}

	return &types.MsgRegisterInterchainQueryResponse{Id: lastID}, nil
}

func (k msgServer) SubmitQueryResult(goCtx context.Context, msg *types.MsgSubmitQueryResult) (*types.MsgSubmitQueryResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	query, err := k.GetQueryByID(ctx, msg.QueryId)
	if err != nil {
		return nil, fmt.Errorf("failed to get query by id: %w", err)
	}

	if msg.Result.KvResults != nil {
		resp, err := k.ibcKeeper.ConnectionConsensusState(goCtx, &ibcconnectiontypes.QueryConnectionConsensusStateRequest{
			ConnectionId:   query.ConnectionId,
			RevisionNumber: 2,
			RevisionHeight: msg.Result.Height + 1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get consensus state: %w", err)
		}

		consensusState, err := ibcclienttypes.UnpackConsensusState(resp.ConsensusState)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack consesus state: %w", err)
		}

		for _, result := range msg.Result.KvResults {
			proof, err := ibccommitmenttypes.ConvertProofs(result.Proof)
			if err != nil {
				return nil, fmt.Errorf("failed to convert crypto.ProofOps to MerkleProof: %w", err)
			}

			path := ibccommitmenttypes.NewMerklePath(result.StoragePrefix, string(result.Key))

			if err := proof.VerifyMembership(ibccommitmenttypes.GetSDKSpecs(), consensusState.GetRoot(), path, result.Value); err != nil {
				return nil, fmt.Errorf("failed to verify proof: %w", err)
			}
		}
	}

	for _, block := range msg.Result.Blocks {
		tmHeader, tmNextHeader, err := k.unpackAndVerifyHeaders(ctx, msg.ClientId, block)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack and verify headers: %w", err)
		}

		for _, tx := range block.Txs {
			if err = verifyTransaction(tmHeader, tmNextHeader, tx); err != nil {
				return nil, fmt.Errorf("failed to verify transaction %s: %w", hex.EncodeToString(tmtypes.Tx(tx.Data).Hash()), err)
			}
		}
	}

	if err = k.SaveQueryResult(ctx, msg.QueryId, msg.Result); err != nil {
		return nil, fmt.Errorf("failed to save query result: %w", err)
	}

	return &types.MsgSubmitQueryResultResponse{}, nil
}

var _ types.MsgServer = msgServer{}
