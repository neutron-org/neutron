package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

type MsgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

var _ types.MsgServer = MsgServer{}

func (k MsgServer) Deposit(
	goCtx context.Context,
	msg *types.MsgDeposit,
) (*types.MsgDepositResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgDeposit")
	}

	if err := k.AssertNotPaused(goCtx); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)

	pairID, err := types.NewPairID(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}

	// sort amounts
	amounts0, amounts1 := SortAmounts(msg.TokenA, pairID.Token0, msg.AmountsA, msg.AmountsB)

	tickIndexes := NormalizeAllTickIndexes(msg.TokenA, pairID.Token0, msg.TickIndexesAToB)

	Amounts0Deposit, Amounts1Deposit, sharesIssued, failedDeposits, err := k.DepositCore(
		goCtx,
		pairID,
		callerAddr,
		receiverAddr,
		amounts0,
		amounts1,
		tickIndexes,
		msg.Fees,
		msg.Options,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgDepositResponse{
		Reserve0Deposited: Amounts0Deposit,
		Reserve1Deposited: Amounts1Deposit,
		FailedDeposits:    failedDeposits,
		SharesIssued:      sharesIssued,
	}, nil
}

func (k MsgServer) Withdrawal(
	goCtx context.Context,
	msg *types.MsgWithdrawal,
) (*types.MsgWithdrawalResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgWithdrawal")
	}

	if err := k.AssertNotPaused(goCtx); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)

	pairID, err := types.NewPairID(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}

	tickIndexes := NormalizeAllTickIndexes(msg.TokenA, pairID.Token0, msg.TickIndexesAToB)

	reserve0ToRemoved, reserve1ToRemoved, sharesBurned, err := k.WithdrawCore(
		goCtx,
		pairID,
		callerAddr,
		receiverAddr,
		msg.SharesToRemove,
		tickIndexes,
		msg.Fees,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawalResponse{
		Reserve0Withdrawn: reserve0ToRemoved,
		Reserve1Withdrawn: reserve1ToRemoved,
		SharesBurned:      sharesBurned,
	}, nil
}

func (k MsgServer) PlaceLimitOrder(
	goCtx context.Context,
	msg *types.MsgPlaceLimitOrder,
) (*types.MsgPlaceLimitOrderResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgPlaceLimitOrder")
	}

	if err := k.AssertNotPaused(goCtx); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)

	err := msg.ValidateGoodTilExpiration(ctx.BlockTime())
	if err != nil {
		return &types.MsgPlaceLimitOrderResponse{}, err
	}
	tickIndex := msg.TickIndexInToOut
	if msg.LimitSellPrice != nil {
		limitBuyPrice := math_utils.OnePrecDec().Quo(*msg.LimitSellPrice)
		tickIndex, err = types.CalcTickIndexFromPrice(limitBuyPrice)
		if err != nil {
			return &types.MsgPlaceLimitOrderResponse{}, errors.Wrapf(err, "invalid LimitSellPrice %s", msg.LimitSellPrice.String())
		}
	}
	trancheKey, coinIn, swapInCoin, coinOutSwap, err := k.PlaceLimitOrderCore(
		goCtx,
		msg.TokenIn,
		msg.TokenOut,
		msg.AmountIn,
		tickIndex,
		msg.OrderType,
		msg.ExpirationTime,
		msg.MaxAmountOut,
		msg.MinAverageSellPrice,
		callerAddr,
		receiverAddr,
	)
	if err != nil {
		return &types.MsgPlaceLimitOrderResponse{}, err
	}

	return &types.MsgPlaceLimitOrderResponse{
		TrancheKey:   trancheKey,
		CoinIn:       coinIn,
		TakerCoinOut: coinOutSwap,
		TakerCoinIn:  swapInCoin,
	}, nil
}

func (k MsgServer) WithdrawFilledLimitOrder(
	goCtx context.Context,
	msg *types.MsgWithdrawFilledLimitOrder,
) (*types.MsgWithdrawFilledLimitOrderResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgWithdrawFilledLimitOrder")
	}

	if err := k.AssertNotPaused(goCtx); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)

	takerCoinOut, makerCoinOut, err := k.WithdrawFilledLimitOrderCore(
		goCtx,
		msg.TrancheKey,
		callerAddr,
	)
	if err != nil {
		return &types.MsgWithdrawFilledLimitOrderResponse{}, err
	}

	return &types.MsgWithdrawFilledLimitOrderResponse{
		TakerCoinOut: takerCoinOut,
		MakerCoinOut: makerCoinOut,
	}, nil
}

func (k MsgServer) CancelLimitOrder(
	goCtx context.Context,
	msg *types.MsgCancelLimitOrder,
) (*types.MsgCancelLimitOrderResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgCancelLimitOrder")
	}

	if err := k.AssertNotPaused(goCtx); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)

	makerCoinOut, takerCoinOut, err := k.CancelLimitOrderCore(
		goCtx,
		msg.TrancheKey,
		callerAddr,
	)
	if err != nil {
		return &types.MsgCancelLimitOrderResponse{}, err
	}

	return &types.MsgCancelLimitOrderResponse{
		TakerCoinOut: takerCoinOut,
		MakerCoinOut: makerCoinOut,
	}, nil
}

func (k MsgServer) MultiHopSwap(
	goCtx context.Context,
	msg *types.MsgMultiHopSwap,
) (*types.MsgMultiHopSwapResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgMultiHopSwap")
	}

	if err := k.AssertNotPaused(goCtx); err != nil {
		return nil, err
	}

	callerAddr := sdk.MustAccAddressFromBech32(msg.Creator)
	receiverAddr := sdk.MustAccAddressFromBech32(msg.Receiver)

	coinOut, route, dust, err := k.MultiHopSwapCore(
		goCtx,
		msg.AmountIn,
		msg.Routes,
		msg.ExitLimitPrice,
		msg.PickBestRoute,
		callerAddr,
		receiverAddr,
	)
	if err != nil {
		return &types.MsgMultiHopSwapResponse{}, err
	}
	return &types.MsgMultiHopSwapResponse{
		CoinOut: coinOut,
		Route:   &types.MultiHopRoute{Hops: route},
		Dust:    dust,
	}, nil
}

func (k MsgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgUpdateParams")
	}

	authority := k.GetAuthority()
	if authority != req.Authority {
		return nil, status.Errorf(codes.PermissionDenied, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k MsgServer) AssertNotPaused(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	paused := k.GetParams(ctx).Paused

	if paused {
		return types.ErrDexPaused
	}
	return nil
}
