package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreateDenom(goCtx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgCreateDenom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	denom, err := server.Keeper.CreateDenom(ctx, msg.Sender, msg.Subdenom)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgCreateDenom,
			sdk.NewAttribute(types.AttributeCreator, msg.Sender),
			sdk.NewAttribute(types.AttributeNewTokenDenom, denom),
		),
	})

	return &types.MsgCreateDenomResponse{
		NewTokenDenom: denom,
	}, nil
}

func (server msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgMint")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// pay some extra gas cost to give a better error here.
	_, denomExists := server.bankKeeper.GetDenomMetaData(ctx, msg.Amount.Denom)
	if !denomExists {
		return nil, types.ErrDenomDoesNotExist.Wrapf("denom: %s", msg.Amount.Denom)
	}

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	if msg.MintToAddress == "" {
		msg.MintToAddress = msg.Sender
	}

	err = server.Keeper.mintTo(ctx, msg.Amount, msg.MintToAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgMint,
			sdk.NewAttribute(types.AttributeMintToAddress, msg.MintToAddress),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
		),
	})

	return &types.MsgMintResponse{}, nil
}

func (server msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgBurn")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	if msg.BurnFromAddress == "" {
		msg.BurnFromAddress = msg.Sender
	}

	accountI := server.Keeper.accountKeeper.GetAccount(ctx, sdk.AccAddress(msg.BurnFromAddress))
	_, ok := accountI.(sdk.ModuleAccountI)
	if ok {
		return nil, types.ErrBurnFromModuleAccount
	}

	err = server.Keeper.burnFrom(ctx, msg.Amount, msg.BurnFromAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgBurn,
			sdk.NewAttribute(types.AttributeBurnFromAddress, msg.BurnFromAddress),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
		),
	})

	return &types.MsgBurnResponse{}, nil
}

func (server msgServer) ForceTransfer(goCtx context.Context, msg *types.MsgForceTransfer) (*types.MsgForceTransferResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgForceTransfer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	err = server.Keeper.forceTransfer(ctx, msg.Amount, msg.TransferFromAddress, msg.TransferToAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgForceTransfer,
			sdk.NewAttribute(types.AttributeTransferFromAddress, msg.TransferFromAddress),
			sdk.NewAttribute(types.AttributeTransferToAddress, msg.TransferToAddress),
			sdk.NewAttribute(types.AttributeAmount, msg.Amount.String()),
		),
	})

	return &types.MsgForceTransferResponse{}, nil
}

func (server msgServer) ChangeAdmin(goCtx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgChangeAdmin")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized.Wrapf("need: %s, received: %s, denom: %s", authorityMetadata.GetAdmin(), msg.Sender, msg.Denom)
	}

	err = server.Keeper.setAdmin(ctx, msg.Denom, msg.NewAdmin)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgChangeAdmin,
			sdk.NewAttribute(types.AttributeDenom, msg.GetDenom()),
			sdk.NewAttribute(types.AttributeNewAdmin, msg.NewAdmin),
		),
	})

	return &types.MsgChangeAdminResponse{}, nil
}

func (server msgServer) SetDenomMetadata(goCtx context.Context, msg *types.MsgSetDenomMetadata) (*types.MsgSetDenomMetadataResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgSetDenomMetadata")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Defense in depth validation of metadata
	err := msg.Metadata.Validate()
	if err != nil {
		return nil, err
	}

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Metadata.Base)
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	server.Keeper.bankKeeper.SetDenomMetaData(ctx, msg.Metadata)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetDenomMetadata,
			sdk.NewAttribute(types.AttributeDenom, msg.Metadata.Base),
			sdk.NewAttribute(types.AttributeDenomMetadata, msg.Metadata.String()),
		),
	})

	return &types.MsgSetDenomMetadataResponse{}, nil
}

func (server msgServer) SetBeforeSendHook(goCtx context.Context, msg *types.MsgSetBeforeSendHook) (*types.MsgSetBeforeSendHookResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgSetBeforeSendHook")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	// If we are not removing a hook make sure it has been already whitelisted
	if msg.ContractAddr != "" {
		// msg.ContractAddr has already been validated
		cwAddr := sdk.MustAccAddressFromBech32(msg.ContractAddr)
		if err := server.Keeper.AssertIsHookWhitelisted(ctx, msg.Denom, cwAddr); err != nil {
			return nil, err
		}
	}

	err = server.Keeper.setBeforeSendHook(ctx, msg.Denom, msg.ContractAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetBeforeSendHook,
			sdk.NewAttribute(types.AttributeDenom, msg.GetDenom()),
			sdk.NewAttribute(types.AttributeBeforeSendHookAddress, msg.GetContractAddr()),
		),
	})

	return &types.MsgSetBeforeSendHookResponse{}, nil
}

// UpdateParams updates the module parameters
func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgUpdateParams")
	}

	authority := k.GetAuthority()
	if authority != req.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
