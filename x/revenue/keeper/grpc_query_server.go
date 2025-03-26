package keeper

import (
	"context"
	"errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
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

func (s queryServer) PaymentInfo(goCtx context.Context, request *revenuetypes.QueryPaymentInfoRequest) (*revenuetypes.QueryPaymentInfoResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)

	ps, err := s.keeper.getPaymentSchedule(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get payment schedule: %s", err)
	}
	psi, err := ps.IntoPaymentScheduleI()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid payment schedule: %s", err)
	}

	twap, err := s.keeper.GetTWAP(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calc TWAP: %s", err)
	}

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get module params: %s", err)
	}
	baseRevenueAmount, err := s.keeper.CalcBaseRevenueAmount(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calc base revenue amount: %s", err)
	}

	return &revenuetypes.QueryPaymentInfoResponse{
		PaymentSchedule:         *ps,
		EffectivePeriodProgress: psi.EffectivePeriodProgress(ctx),
		RewardAssetTwap:         twap,
		BaseRevenueAmount:       sdktypes.NewCoin(params.RewardAsset, baseRevenueAmount),
	}, nil
}

func (s queryServer) ValidatorStats(goCtx context.Context, request *revenuetypes.QueryValidatorStatsRequest) (*revenuetypes.QueryValidatorStatsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)

	valOperAddr, err := sdktypes.ValAddressFromBech32(request.ValOperAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	valInfo, err := s.keeper.GetValidatorInfo(ctx, valOperAddr)
	if err != nil {
		if errors.Is(err, revenuetypes.ErrNoValidatorInfoFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get module params: %s", err)
	}

	ps, err := s.keeper.GetPaymentScheduleI(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get payment schedule: %s", err)
	}

	baseRevenueAmount, err := s.keeper.CalcBaseRevenueAmount(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	totalBlocksInPeriod := ps.TotalBlocksInPeriod(ctx)

	pr, valCompensation := evaluateValCommitment(
		params,
		baseRevenueAmount,
		valInfo,
		totalBlocksInPeriod,
	)

	return &revenuetypes.QueryValidatorStatsResponse{
		Stats: revenuetypes.ValidatorStats{
			ValidatorInfo:               valInfo,
			TotalProducedBlocksInPeriod: totalBlocksInPeriod,
			PerformanceRating:           pr,
			ExpectedRevenue:             sdktypes.NewCoin(params.RewardAsset, valCompensation),
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

	ps, err := s.keeper.GetPaymentScheduleI(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get payment schedule: %s", err)
	}

	baseRevenueAmount, err := s.keeper.CalcBaseRevenueAmount(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	totalBlocksInPeriod := ps.TotalBlocksInPeriod(ctx)

	valStats := make([]revenuetypes.ValidatorStats, 0, len(valsInfo))
	for _, valInfo := range valsInfo {
		pr, valCompensation := evaluateValCommitment(
			params,
			baseRevenueAmount,
			valInfo,
			totalBlocksInPeriod,
		)

		valStats = append(valStats, revenuetypes.ValidatorStats{
			ValidatorInfo:               valInfo,
			TotalProducedBlocksInPeriod: totalBlocksInPeriod,
			PerformanceRating:           pr,
			ExpectedRevenue:             sdktypes.NewCoin(params.RewardAsset, valCompensation),
		})
	}

	return &revenuetypes.QueryValidatorsStatsResponse{
		Stats: valStats,
	}, nil
}
