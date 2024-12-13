package keeper

import (
	"context"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (s msgServer) UpdateParams(context context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// TODO: implement
	return nil, nil
}
