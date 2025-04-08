package types

import (
	"context"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"google.golang.org/grpc"

	feerefundertypes "github.com/neutron-org/neutron/v6/x/feerefunder/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg *MsgTransfer) Validate(isContract bool) error {
	if isContract {
		if err := msg.Fee.Validate(); err != nil {
			return err
		}
	}

	sdkMsg := types.NewMsgTransfer(msg.SourcePort, msg.SourceChannel, msg.Token, msg.Sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp, msg.Memo)
	return sdkMsg.ValidateBasic()
}

func (msg *MsgTransfer) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{fromAddress}
}

// MsgOrigTransferHandler - 1) helps to bind `/neutron.transfer.Msg/Transfer` as a handler for `ibc.applications.transfer.v1.MsgTransfer`
// 2) converts `ibc.applications.transfer.v1.MsgTransfer` into `neutron.transfer.MsgTransfer` before processing.
//
//nolint:revive // we cant rearrange arguments since we need to meet the type requirement
func MsgOrigTransferHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.MsgTransfer)
	if err := dec(in); err != nil {
		return nil, err
	}
	conv := &MsgTransfer{
		SourcePort:       in.SourcePort,
		SourceChannel:    in.SourceChannel,
		Token:            in.Token,
		Sender:           in.Sender,
		Receiver:         in.Receiver,
		TimeoutHeight:    in.TimeoutHeight,
		TimeoutTimestamp: in.TimeoutTimestamp,
		Memo:             in.Memo,
		Fee:              feerefundertypes.Fee{},
	}
	if interceptor == nil {
		return srv.(MsgServer).Transfer(ctx, conv)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/neutron.transfer.Msg/Transfer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		reqT := req.(*types.MsgTransfer)
		convReq := &MsgTransfer{
			SourcePort:       reqT.SourcePort,
			SourceChannel:    reqT.SourceChannel,
			Token:            reqT.Token,
			Sender:           reqT.Sender,
			Receiver:         reqT.Receiver,
			TimeoutHeight:    reqT.TimeoutHeight,
			TimeoutTimestamp: reqT.TimeoutTimestamp,
			Memo:             reqT.Memo,
			Fee:              feerefundertypes.Fee{},
		}
		return srv.(MsgServer).Transfer(ctx, convReq)
	}
	return interceptor(ctx, conv, info, handler)
}

// MsgUpdateParamsHandler handler helps to bind `/ibc.applications.transfer.v1.Msg/UpdateParams`.
//
//nolint:revive // we cant rearrange arguments since we need to meet the type requirement
func MsgUpdateParamsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	conv := &MsgUpdateParams{
		Signer: in.Signer,
		Params: in.Params,
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, conv)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ibc.applications.transfer.v1.Msg/UpdateParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		reqT := req.(*types.MsgUpdateParams)
		convReq := &MsgUpdateParams{
			Signer: reqT.Signer,
			Params: reqT.Params,
		}
		return srv.(MsgServer).UpdateParams(ctx, convReq)
	}
	return interceptor(ctx, in, info, handler)
}

var MsgServiceDescOrig = grpc.ServiceDesc{
	ServiceName: "ibc.applications.transfer.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Transfer",
			Handler:    MsgOrigTransferHandler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    MsgUpdateParamsHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ibc/applications/transfer/v1/tx.proto",
}
