package keeper

import (
	"context"

	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) revenuetypes.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ revenuetypes.MsgServer = msgServer{}

func (s msgServer) UpdateParams(context context.Context, msg *revenuetypes.MsgUpdateParams) (*revenuetypes.MsgUpdateParamsResponse, error) {
	// TODO: implement
	return nil, nil
}
