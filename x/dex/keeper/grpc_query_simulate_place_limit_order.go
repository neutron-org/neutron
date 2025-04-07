package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func (k Keeper) SimulatePlaceLimitOrder(
	goCtx context.Context,
	req *types.QuerySimulatePlaceLimitOrderRequest,
) (*types.QuerySimulatePlaceLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, _ := ctx.CacheContext()

	msg := req.Msg
	msg.Creator = types.DummyAddress
	msg.Receiver = types.DummyAddress

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	err := msg.ValidateGoodTilExpiration(ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)
	takerTradePairID, err := types.NewTradePairID(msg.TokenIn, msg.TokenOut)
	if err != nil {
		return nil, err
	}
	tickIndex := msg.TickIndexInToOut
	if msg.LimitSellPrice != nil {
		limitBuyPrice := math_utils.OnePrecDec().Quo(*msg.LimitSellPrice)
		tickIndex, err = types.CalcTickIndexFromPrice(limitBuyPrice)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid LimitSellPrice %s", msg.LimitSellPrice.String())
		}
	}
	trancheKey, totalIn, takerCoinIn, takerCoinOut, _, _, err := k.ExecutePlaceLimitOrder(
		cacheCtx,
		takerTradePairID,
		msg.AmountIn,
		tickIndex,
		msg.OrderType,
		msg.ExpirationTime,
		msg.MaxAmountOut,
		msg.MinAverageSellPrice,
		receiverAddr,
	)
	if err != nil {
		return nil, err
	}

	coinIn := sdk.NewCoin(msg.TokenIn, totalIn)
	return &types.QuerySimulatePlaceLimitOrderResponse{
		Resp: &types.MsgPlaceLimitOrderResponse{
			TrancheKey:   trancheKey,
			CoinIn:       coinIn,
			TakerCoinIn:  takerCoinIn,
			TakerCoinOut: takerCoinOut,
		},
	}, nil
}
