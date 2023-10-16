package types_test

import (
	"testing"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite
	Ctx sdk.Context
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Ctx = testutil.DefaultContext(
		sdk.NewKVStoreKey(types.StoreKey),
		sdk.NewTransientStoreKey("transient_test"),
	)
}

func dummyAfterEpochEndEvent(epochIdentifier string) sdk.Event {
	return sdk.NewEvent(
		"afterEpochEnd",
		sdk.NewAttribute("epochIdentifier", epochIdentifier),
	)
}

func dummyBeforeEpochStartEvent(epochIdentifier string) sdk.Event {
	return sdk.NewEvent(
		"beforeEpochStart",
		sdk.NewAttribute("epochIdentifier", epochIdentifier),
	)
}

var errDummy = errors.New("9", 9, "dummyError")

// dummyEpochHook is a struct satisfying the epoch hook interface,
// that maintains a counter for how many times its been successfully called,
// and a boolean for whether it should panic during its execution.
type dummyEpochHook struct {
	successCounter int
	shouldPanic    bool
	shouldError    bool
}

func (hook *dummyEpochHook) AfterEpochEnd(
	ctx sdk.Context,
	epochIdentifier string,
) error {
	if hook.shouldPanic {
		panic("dummyEpochHook is panicking")
	}
	if hook.shouldError {
		return errDummy
	}
	hook.successCounter++
	ctx.EventManager().EmitEvent(dummyAfterEpochEndEvent(epochIdentifier))

	return nil
}

func (hook *dummyEpochHook) BeforeEpochStart(
	ctx sdk.Context,
	epochIdentifier string,
) error {
	if hook.shouldPanic {
		panic("dummyEpochHook is panicking")
	}
	if hook.shouldError {
		return errDummy
	}
	hook.successCounter++
	ctx.EventManager().EmitEvent(dummyBeforeEpochStartEvent(epochIdentifier))

	return nil
}

func (hook *dummyEpochHook) Clone() *dummyEpochHook {
	newHook := dummyEpochHook{
		shouldPanic:    hook.shouldPanic,
		successCounter: hook.successCounter,
		shouldError:    hook.shouldError,
	}
	return &newHook
}

var _ types.EpochHooks = &dummyEpochHook{}

func (s *KeeperTestSuite) TestHooksPanicRecovery() {
	// panicHook := dummyEpochHook{shouldPanic: true}
	noPanicHook := dummyEpochHook{shouldPanic: false}
	// errorHook := dummyEpochHook{shouldError: true}
	// noErrorHook := dummyEpochHook{shouldError: false} // same as nopanic
	// simpleHooks := []dummyEpochHook{panicHook, noPanicHook, errorHook, noErrorHook}

	tests := []struct {
		hooks                 []dummyEpochHook
		expectedCounterValues []int
		lenEvents             int
	}{
		{[]dummyEpochHook{noPanicHook}, []int{1}, 1},
		// {[]dummyEpochHook{panicHook}, []int{0}, 0},
		// {[]dummyEpochHook{errorHook}, []int{0}, 0},
		// {simpleHooks, []int{0, 1, 0, 1}, 2},
	}

	for tcIndex, tc := range tests {
		for epochActionSelector := 0; epochActionSelector < 2; epochActionSelector++ {
			s.SetupTest()
			hookRefs := []types.EpochHooks{}

			for _, hook := range tc.hooks {
				hookRefs = append(hookRefs, hook.Clone())
			}

			hooks := types.NewMultiEpochHooks(hookRefs...)

			events := func(epochID string, dummyEvent func(id string) sdk.Event) sdk.Events {
				evts := make(sdk.Events, tc.lenEvents)
				for i := 0; i < tc.lenEvents; i++ {
					evts[i] = dummyEvent(epochID)
				}

				return evts
			}

			s.NotPanics(func() {
				if epochActionSelector == 0 {
					err := hooks.BeforeEpochStart(s.Ctx, "id")
					s.Require().NoError(err)
					s.Require().Equal(
						events(
							"id",
							dummyBeforeEpochStartEvent,
						),
						s.Ctx.EventManager().Events(),
						"test case index %d, before epoch event check", tcIndex,
					)
				} else if epochActionSelector == 1 {
					err := hooks.AfterEpochEnd(s.Ctx, "id")
					s.Require().NoError(err)
					s.Require().Equal(events("id", dummyAfterEpochEndEvent), s.Ctx.EventManager().Events(),
						"test case index %d, after epoch event check", tcIndex)
				}
			})

			for i := 0; i < len(hooks); i++ {
				epochHook := hookRefs[i].(*dummyEpochHook)
				s.Require().
					Equal(tc.expectedCounterValues[i], epochHook.successCounter, "test case index %d", tcIndex)
			}
		}
	}
}
