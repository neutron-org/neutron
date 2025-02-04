package keeper

import (
	"context"
	"errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper *Keeper) revenuetypes.QueryServer {
	return &queryServer{keeper: keeper}
}

var _ revenuetypes.QueryServer = queryServer{}

func (s queryServer) Params(goCtx context.Context, request *revenuetypes.QueryParamsRequest) (*revenuetypes.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)
	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &revenuetypes.QueryParamsResponse{Params: params}, nil
}

func (s queryServer) State(goCtx context.Context, request *revenuetypes.QueryStateRequest) (*revenuetypes.QueryStateResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)
	state, err := s.keeper.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &revenuetypes.QueryStateResponse{State: state}, nil
}

func (s queryServer) ValidatorStats(goCtx context.Context, request *revenuetypes.QueryValidatorStatsRequest) (*revenuetypes.QueryValidatorStatsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)

	ca, err := sdktypes.ConsAddressFromBech32(request.ConsensusAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	valInfo, err := s.keeper.GetValidatorInfo(ctx, ca)
	if err != nil {
		if errors.Is(err, revenuetypes.ErrNoValidatorInfoFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	state, err := s.keeper.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pscv := state.PaymentSchedule.GetCachedValue()
	ps, ok := pscv.(revenuetypes.PaymentSchedule)
	if !ok {
		return nil, status.Errorf(codes.Internal, "expected state.PaymentSchedule to be of type PaymentSchedule, but got %T", pscv)
	}

	blocksPerPeriod := ps.TotalBlocksInPeriod(ctx)
	pr := PerformanceRating(
		params.BlocksPerformanceRequirement,
		params.OracleVotesPerformanceRequirement,
		int64(blocksPerPeriod-valInfo.GetCommitedBlocksInPeriod()),
		int64(blocksPerPeriod-valInfo.GetCommitedOracleVotesInPeriod()),
		int64(blocksPerPeriod),
	)

	amount, err := s.keeper.CalcBaseRevenueAmount(ctx, params.BaseCompensation)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &revenuetypes.QueryValidatorStatsResponse{
		Stats: revenuetypes.ValidatorStats{
			ValidatorInfo:     valInfo,
			PerformanceRating: pr,
			ExpectedRevenue:   pr.MulInt(amount).TruncateInt(),
		},
	}, nil
}

func (s queryServer) ValidatorsStats(goCtx context.Context, request *revenuetypes.QueryValidatorsStatsRequest) (*revenuetypes.QueryValidatorsStatsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)

	valsInfo, err := s.keeper.GetAllValidatorInfo(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	state, err := s.keeper.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pscv := state.PaymentSchedule.GetCachedValue()
	ps, ok := pscv.(revenuetypes.PaymentSchedule)
	if !ok {
		return nil, status.Errorf(codes.Internal, "expected state.PaymentSchedule to be of type PaymentSchedule, but got %T", pscv)
	}

	amount, err := s.keeper.CalcBaseRevenueAmount(ctx, params.BaseCompensation)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	valStats := make([]revenuetypes.ValidatorStats, 0, len(valsInfo))
	for _, valInfo := range valsInfo {
		blocksPerPeriod := ps.TotalBlocksInPeriod(ctx)
		pr := PerformanceRating(
			params.BlocksPerformanceRequirement,
			params.OracleVotesPerformanceRequirement,
			int64(blocksPerPeriod-valInfo.GetCommitedBlocksInPeriod()),
			int64(blocksPerPeriod-valInfo.GetCommitedOracleVotesInPeriod()),
			int64(blocksPerPeriod),
		)
		valStats = append(valStats, revenuetypes.ValidatorStats{
			ValidatorInfo:     valInfo,
			PerformanceRating: pr,
			ExpectedRevenue:   pr.MulInt(amount).TruncateInt(),
		})
	}

	return &revenuetypes.QueryValidatorsStatsResponse{
		Stats: valStats,
	}, nil
}
