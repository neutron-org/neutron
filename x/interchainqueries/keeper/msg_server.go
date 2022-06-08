package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
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
		ConnectionId:     msg.ConnectionId,
		LastLocalHeight:  uint64(ctx.BlockHeight()),
		LastRemoteHeight: 0,
	}

	k.SetLastRegisteredQueryKey(ctx, lastID)
	k.SaveQuery(ctx, registeredQuery)

	return &types.MsgRegisterInterchainQueryResponse{Id: lastID}, nil
}

func (k msgServer) SubmitQueryResult(goCtx context.Context, msg *types.MsgSubmitQueryResult) (*types.MsgSubmitQueryResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	query, err := k.GetQueryByID(ctx, msg.QueryId)
	if err != nil {
		return nil, fmt.Errorf("failed to get query by id: %w", err)
	}

	resp, err := k.ibcKeeper.ConnectionConsensusState(goCtx, &ibcconnectiontypes.QueryConnectionConsensusStateRequest{
		ConnectionId:   query.ConnectionId,
		RevisionNumber: 1,
		RevisionHeight: msg.Result.Height,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus state: %w", err)
	}

	consensusState, err := ibcclienttypes.UnpackConsensusState(resp.ConsensusState)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack consesus state: %w", err)
	}

	consensusState.ClientType()

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

	for _, block := range msg.Result.Blocks {
		header, err := ibcclienttypes.UnpackHeader(block.Header)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack block header: %w", err)
		}

		if err = k.ibcKeeper.ClientKeeper.UpdateClient(ctx, msg.ClientId, header); err != nil {
			return nil, fmt.Errorf("failed to vefify header and update client state: %w", err)
		}

		tmHeader, ok := header.(*tendermintLightClientTypes.Header)
		if !ok {
			return nil, fmt.Errorf("failed to cast header to tendermint Header: %w", err)
		}

		for _, tx := range block.Txs {
			inclusionProof, err := merkle.ProofFromProto(tx.InclusionProof)
			if err != nil {
				return nil, fmt.Errorf("failed to convert proto proof to merkle proof: %w", err)
			}

			if err = inclusionProof.Verify(tmHeader.Header.DataHash, tx.Data); err != nil {
				return nil, fmt.Errorf("failed to verify inclusion proof: %w", err)
			}

			deliveryProof, err := merkle.ProofFromProto(tx.DeliveryProof)
			if err != nil {
				return nil, fmt.Errorf("failed to convert proto proof to merkle proof: %w", err)
			}

			responseTx := deterministicResponseDeliverTx(tx.Response)

			responseTxBz, err := responseTx.Marshal()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal ResponseDeliveryTx: %w", err)
			}

			if err = deliveryProof.Verify(tmHeader.Header.LastResultsHash, responseTxBz); err != nil {
				return nil, fmt.Errorf("failed to verify delivery proof: %w", err)
			}

			if deliveryProof.Index != inclusionProof.Index {
				return nil, fmt.Errorf("inclusion proof index and delivery proof index are not equal: %d != %d", inclusionProof.Index, deliveryProof.Index)
			}
		}
	}

	k.SaveQueryResult(ctx, msg.QueryId, msg.Result)

	return &types.MsgSubmitQueryResultResponse{}, nil
}

// deterministicResponseDeliverTx strips non-deterministic fields from
// ResponseDeliverTx and returns another ResponseDeliverTx.
func deterministicResponseDeliverTx(response *abci.ResponseDeliverTx) *abci.ResponseDeliverTx {
	return &abci.ResponseDeliverTx{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}

var _ types.MsgServer = msgServer{}
