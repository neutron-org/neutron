package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) DenomAuthorityMetadata(ctx context.Context, req *types.QueryDenomAuthorityMetadataRequest) (*types.QueryDenomAuthorityMetadataResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	denom := fmt.Sprintf("factory/%s/%s", req.GetCreator(), req.GetSubdenom())
	authorityMetadata, err := k.GetAuthorityMetadata(sdkCtx, denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomAuthorityMetadataResponse{AuthorityMetadata: authorityMetadata}, nil
}

func (k Keeper) DenomsFromCreator(ctx context.Context, req *types.QueryDenomsFromCreatorRequest) (*types.QueryDenomsFromCreatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	denoms := k.getDenomsFromCreator(sdkCtx, req.GetCreator())
	return &types.QueryDenomsFromCreatorResponse{Denoms: denoms}, nil
}

func (k Keeper) BeforeSendHookAddress(ctx context.Context, req *types.QueryBeforeSendHookAddressRequest) (*types.QueryBeforeSendHookAddressResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	denom := fmt.Sprintf("factory/%s/%s", req.GetCreator(), req.GetSubdenom())
	contractAddr := k.GetBeforeSendHook(sdkCtx, denom)

	return &types.QueryBeforeSendHookAddressResponse{ContractAddr: contractAddr}, nil
}

func (k Keeper) FullDenom(_ context.Context, req *types.QueryFullDenomRequest) (*types.QueryFullDenomResponse, error) {
	if _, err := sdk.AccAddressFromBech32(req.Creator); err != nil {
		return nil, err
	}

	fullDenom, err := types.GetTokenDenom(req.Creator, req.Subdenom)
	if err != nil {
		return nil, err
	}

	return &types.QueryFullDenomResponse{FullDenom: fullDenom}, nil
}
