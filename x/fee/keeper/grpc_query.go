package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/x/fee/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) FeeInfo(goCtx context.Context, request *types.FeeInfoRequest) (*types.FeeInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	feeInfo, err := k.GetFeeInfo(ctx, channeltypes.NewPacketId(request.PortId, request.ChannelId, request.Sequence))
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no fee info found for port_id = %s, channel_id=%s, sequnce=%d", request.PortId, request.ChannelId, request.Sequence)
	}

	return &types.FeeInfoResponse{FeeInfo: feeInfo}, nil
}
