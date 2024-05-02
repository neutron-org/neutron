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
	auctionante "github.com/skip-mev/block-sdk/v2/x/auction/ante"
	auctionkeeper "github.com/skip-mev/block-sdk/v2/x/auction/keeper"

	globalfeeante "github.com/neutron-org/neutron/v4/x/globalfee/ante"
	globalfeekeeper "github.com/neutron-org/neutron/v4/x/globalfee/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper             *ibckeeper.Keeper
	ConsumerKeeper        ibcconsumerkeeper.Keeper
	GlobalFeeKeeper       globalfeekeeper.Keeper
	WasmConfig            *wasmTypes.WasmConfig
	TXCounterStoreService corestoretypes.KVStoreService

	// dependencies for the x/auction ante-handler
	AuctionKeeper auctionkeeper.Keeper
	TxEncoder     sdk.TxEncoder
	MEVLane       auctionante.MEVLane
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

	if options.MEVLane == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "mev lane is required for AnteHandler")
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
		globalfeeante.NewFeeDecorator(options.GlobalFeeKeeper),

		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		auctionante.NewAuctionDecorator(
			options.AuctionKeeper,
			options.TxEncoder,
			options.MEVLane,
		),
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
