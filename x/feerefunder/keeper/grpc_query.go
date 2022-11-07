package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/x/feerefunder/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) FeeInfo(goCtx context.Context, request *types.FeeInfoRequest) (*types.FeeInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	feeInfo, err := k.GetFeeInfo(ctx, types.NewPacketID(request.PortId, request.ChannelId, request.Sequence))
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no feerefunder info found for port_id = %s, channel_id=%s, sequence=%d", request.PortId, request.ChannelId, request.Sequence)
	}

	return &types.FeeInfoResponse{FeeInfo: feeInfo}, nil
}
