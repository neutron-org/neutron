package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type queryServer struct{ k Keeper }

// NewQueryServerImpl creates an implementation of the QueryServer interface for the given keeper
func NewQueryServerImpl(k Keeper) govtypes.QueryServer {
	return &queryServer{k}
}

func (qs queryServer) TallyResult(goCtx context.Context, req *govtypes.QueryTallyResultRequest) (*govtypes.QueryTallyResultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	proposal, ok := qs.k.GetProposal(ctx, req.ProposalId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "proposal %d doesn't exist", req.ProposalId)
	}

	var tallyResult govtypes.TallyResult

	switch {
	case proposal.Status == govtypes.StatusDepositPeriod:
		tallyResult = govtypes.EmptyTallyResult()

	case proposal.Status == govtypes.StatusPassed || proposal.Status == govtypes.StatusRejected:
		tallyResult = proposal.FinalTallyResult

	default:
		// proposal is in voting period
		_, _, tallyResult = qs.k.Tally(ctx, proposal) // replace with our custom Tally function
	}

	return &govtypes.QueryTallyResultResponse{Tally: tallyResult}, nil
}

func (qs queryServer) Proposal(goCtx context.Context, req *govtypes.QueryProposalRequest) (*govtypes.QueryProposalResponse, error) {
	return qs.k.Proposal(goCtx, req)
}

func (qs queryServer) Proposals(goCtx context.Context, req *govtypes.QueryProposalsRequest) (*govtypes.QueryProposalsResponse, error) {
	return qs.k.Proposals(goCtx, req)
}

func (qs queryServer) Vote(goCtx context.Context, req *govtypes.QueryVoteRequest) (*govtypes.QueryVoteResponse, error) {
	return qs.k.Vote(goCtx, req)
}

func (qs queryServer) Votes(goCtx context.Context, req *govtypes.QueryVotesRequest) (*govtypes.QueryVotesResponse, error) {
	return qs.k.Votes(goCtx, req)
}

func (qs queryServer) Params(goCtx context.Context, req *govtypes.QueryParamsRequest) (*govtypes.QueryParamsResponse, error) {
	return qs.k.Params(goCtx, req)
}

func (qs queryServer) Deposit(goCtx context.Context, req *govtypes.QueryDepositRequest) (*govtypes.QueryDepositResponse, error) {
	return qs.k.Deposit(goCtx, req)
}

func (qs queryServer) Deposits(goCtx context.Context, req *govtypes.QueryDepositsRequest) (*govtypes.QueryDepositsResponse, error) {
	return qs.k.Deposits(goCtx, req)
}
