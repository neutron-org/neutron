package app

import (
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	consumerante "github.com/cosmos/interchain-security/v5/app/consumer/ante"
	ibcconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper             *ibckeeper.Keeper
	ConsumerKeeper        ibcconsumerkeeper.Keeper
	WasmConfig            *wasmTypes.WasmConfig
	TXCounterStoreService corestoretypes.KVStoreService
	FeeMarketKeeper       feemarketante.FeeMarketKeeper
}

func NewAnteHandler(options HandlerOptions, logger log.Logger) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.WasmConfig == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "wasm config is required for ante builder")
	}
	if options.TXCounterStoreService == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "tx counter store service is required for ante builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		consumerante.NewDisabledModulesDecorator("/cosmos.evidence", "/cosmos.slashing"),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewDecuctFeeDecoratorWithFallback(options),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	// Don't delete it even if IDE tells you so.
	// This constant depends on build tag.
	if !SkipCcvMsgFilter {
		anteDecorators = append(anteDecorators, consumerante.NewMsgFilterDecorator(options.ConsumerKeeper))
	} else {
		logger.Error("WARNING: BUILT WITH skip_ccv_msg_filter. THIS IS NOT A PRODUCTION BUILD")
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// DeductFeeDecoratorWithFallback is a fee ante decorator which switches between default cosmos-sdk FeeDecorator
// and feemarket's one, depending on the `params.Enabled` field feemarket's module.
type DeductFeeDecoratorWithFallback struct {
	feemarketkeeper    feemarketante.FeeMarketKeeper
	feemarketDecorator feemarketante.FeeMarketCheckDecorator
	cosmosDecorator    ante.DeductFeeDecorator
}

func NewDecuctFeeDecoratorWithFallback(options HandlerOptions) DeductFeeDecoratorWithFallback {
	return DeductFeeDecoratorWithFallback{
		feemarketkeeper: options.FeeMarketKeeper,
		feemarketDecorator: feemarketante.NewFeeMarketCheckDecorator(
			options.FeeMarketKeeper,
		),
		cosmosDecorator: ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
	}
}

func (d DeductFeeDecoratorWithFallback) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	params, err := d.feemarketkeeper.GetParams(ctx)
	if err != nil {
		return ctx, err
	}
	if params.Enabled {
		return d.feemarketDecorator.AnteHandle(ctx, tx, simulate, next)
	}
	return d.cosmosDecorator.AnteHandle(ctx, tx, simulate, next)
}
