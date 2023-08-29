package neutronutils

import sdk "github.com/cosmos/cosmos-sdk/types"

// CreateCachedContext creates a cached context for with a limited gas meter.
func CreateCachedContext(ctx sdk.Context, gasLimit uint64) (sdk.Context, func(), sdk.GasMeter) {
	cacheCtx, writeFn := ctx.CacheContext()
	gasMeter := sdk.NewGasMeter(gasLimit)
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)
	return cacheCtx, writeFn, gasMeter
}
