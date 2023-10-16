package keeper

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dextypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/neutron-org/neutron/x/incentives/types"
)

var _ types.QueryServer = QueryServer{}

// QueryServer defines a wrapper around the incentives module keeper providing gRPC method handlers.
type QueryServer struct {
	*Keeper
}

// NewQueryServer creates a new QueryServer struct.
func NewQueryServer(k *Keeper) QueryServer {
	return QueryServer{Keeper: k}
}

func (q QueryServer) GetModuleStatus(
	goCtx context.Context,
	req *types.GetModuleStatusRequest,
) (*types.GetModuleStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.GetModuleStatusResponse{
		RewardCoins: q.Keeper.GetModuleCoinsToBeDistributed(ctx),
		StakedCoins: q.Keeper.GetModuleStakedCoins(ctx),
		Params:      q.Keeper.GetParams(ctx),
	}, nil
}

func (q QueryServer) GetGaugeByID(
	goCtx context.Context,
	req *types.GetGaugeByIDRequest,
) (*types.GetGaugeByIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	gauge, err := q.Keeper.GetGaugeByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.GetGaugeByIDResponse{Gauge: gauge}, nil
}

func (q QueryServer) GetGaugeQualifyingValue(
	goCtx context.Context,
	req *types.GetGaugeQualifyingValueRequest,
) (*types.GetGaugeQualifyingValueResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	value, err := q.Keeper.GetGaugeQualifyingValue(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.GetGaugeQualifyingValueResponse{QualifyingValue: value}, nil
}

func (q QueryServer) GetGauges(
	goCtx context.Context,
	req *types.GetGaugesRequest,
) (*types.GetGaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var prefix []byte
	switch req.Status {
	case types.GaugeStatus_ACTIVE_UPCOMING:
		prefix = types.KeyPrefixGaugeIndex
	case types.GaugeStatus_ACTIVE:
		prefix = types.KeyPrefixGaugeIndexActive
	case types.GaugeStatus_UPCOMING:
		prefix = types.KeyPrefixGaugeIndexUpcoming
	case types.GaugeStatus_FINISHED:
		prefix = types.KeyPrefixGaugeIndexFinished
	default:
		return nil, sdkerrors.Wrap(types.ErrInvalidRequest, "invalid status filter value")
	}

	var lowerTick, upperTick int64
	var poolMetadata *dextypes.PoolMetadata
	if req.Denom != "" {
		poolMetadata, err := q.dk.GetPoolMetadataByDenom(ctx, req.Denom)
		if err != nil {
			return nil, err
		}
		lowerTick = poolMetadata.Tick - int64(poolMetadata.Fee)
		upperTick = poolMetadata.Tick + int64(poolMetadata.Fee)
	}

	gauges := types.Gauges{}
	store := ctx.KVStore(q.Keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// this may return multiple gauges at once if two gauges start at the same time.
		// for now this is treated as an edge case that is not of importance
		newGauges, err := q.getGaugeFromIDJsonBytes(ctx, iterator.Value())
		if err != nil {
			return nil, err
		}
		if req.Denom != "" {
			for _, gauge := range newGauges {
				if *gauge.DistributeTo.PairID != *poolMetadata.PairID {
					continue
				}
				lowerTickInRange := gauge.DistributeTo.StartTick <= lowerTick &&
					lowerTick <= gauge.DistributeTo.EndTick
				upperTickInRange := gauge.DistributeTo.StartTick <= upperTick &&
					upperTick <= gauge.DistributeTo.EndTick
				if !lowerTickInRange || !upperTickInRange {
					continue
				}
				gauges = append(gauges, gauge)
			}
		} else {
			gauges = append(gauges, newGauges...)
		}
	}

	return &types.GetGaugesResponse{
		Gauges: gauges,
	}, nil
}

func (q QueryServer) GetStakeByID(
	goCtx context.Context,
	req *types.GetStakeByIDRequest,
) (*types.GetStakeByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	stake, err := q.Keeper.GetStakeByID(ctx, req.StakeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.GetStakeByIDResponse{Stake: stake}, nil
}

func (q QueryServer) GetStakes(
	goCtx context.Context,
	req *types.GetStakesRequest,
) (*types.GetStakesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	hasOwner := len(req.Owner) > 0
	if !hasOwner {
		// TODO: Verify this protection is necessary
		return nil, status.Error(
			codes.InvalidArgument,
			"for performance reasons will not return all stakes",
		)
	}

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	stakes := q.Keeper.getStakesByAccount(ctx, owner)
	return &types.GetStakesResponse{
		Stakes: stakes,
	}, nil
}

func (q QueryServer) GetFutureRewardEstimate(
	goCtx context.Context,
	req *types.GetFutureRewardEstimateRequest,
) (*types.GetFutureRewardEstimateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 && len(req.StakeIds) == 0 {
		return nil, sdkerrors.Wrap(types.ErrInvalidRequest, "empty owner")
	}

	if req.NumEpochs > 365 {
		return nil, sdkerrors.Wrap(types.ErrInvalidRequest, "end epoch out of ranges")
	}

	var ownerAddress sdk.AccAddress
	if len(req.Owner) != 0 {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, err
		}
		ownerAddress = owner
	}

	stakes := make(types.Stakes, 0, len(req.StakeIds))
	for _, stakeID := range req.StakeIds {
		stake, err := q.Keeper.GetStakeByID(ctx, stakeID)
		if err != nil {
			return nil, err
		}
		stakes = append(stakes, stake)
	}

	rewards, err := q.Keeper.GetRewardsEstimate(ctx, ownerAddress, stakes, req.NumEpochs)
	if err != nil {
		return nil, err
	}
	return &types.GetFutureRewardEstimateResponse{Coins: rewards}, nil
}

func (q QueryServer) GetAccountHistory(
	goCtx context.Context,
	req *types.GetAccountHistoryRequest,
) (*types.GetAccountHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	accountHistory, found := q.Keeper.GetAccountHistory(ctx, req.Account)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			"Could not locate an account history with that address",
		)
	}
	return &types.GetAccountHistoryResponse{Coins: accountHistory.Coins}, nil
}

// getGaugeFromIDJsonBytes returns gauges from the json bytes of gaugeIDs.
func (q QueryServer) getGaugeFromIDJsonBytes(
	ctx sdk.Context,
	refValue []byte,
) (types.Gauges, error) {
	gauges := types.Gauges{}
	gaugeIDs := []uint64{}

	err := json.Unmarshal(refValue, &gaugeIDs)
	if err != nil {
		return gauges, err
	}

	for _, gaugeID := range gaugeIDs {
		gauge, err := q.Keeper.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			return types.Gauges{}, err
		}

		gauges = append(gauges, gauge)
	}

	return gauges, nil
}
