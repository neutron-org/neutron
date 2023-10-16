package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/utils"
)

type EpochHooks interface {
	// the first block whose timestamp is after the duration is counted as the end of the epoch
	AfterEpochEnd(ctx sdk.Context, epochIdentifier string) error
	// new epoch is next block of epoch end block
	BeforeEpochStart(ctx sdk.Context, epochIdentifier string) error
}

var _ EpochHooks = MultiEpochHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiEpochHooks []EpochHooks

func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the number of epoch that is ending.
func (h MultiEpochHooks) AfterEpochEnd(
	ctx sdk.Context,
	epochIdentifier string,
) error {
	for i := range h {
		panicCatchingEpochHook(ctx, h[i].AfterEpochEnd, epochIdentifier)
	}

	return nil
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is the number of epoch that is starting.
func (h MultiEpochHooks) BeforeEpochStart(
	ctx sdk.Context,
	epochIdentifier string,
) error {
	for i := range h {
		panicCatchingEpochHook(ctx, h[i].BeforeEpochStart, epochIdentifier)
	}

	return nil
}

func panicCatchingEpochHook(
	ctx sdk.Context,
	hookFn func(ctx sdk.Context, epochIdentifier string) error,
	epochIdentifier string,
) {
	wrappedHookFn := func(ctx sdk.Context) error {
		return hookFn(ctx, epochIdentifier)
	}
	// TODO: Thread info for which hook this is, may be dependent on larger hook system refactoring
	err := utils.ApplyFuncIfNoError(ctx, wrappedHookFn)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("error in epoch hook %v", err))
	}
}
