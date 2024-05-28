package app

import (
	"encoding/json"

	"cosmossdk.io/math"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

var FeeDenom = "untrn"

// GenesisState is the genesis state of the blockchain represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	genesisState := ModuleBasics.DefaultGenesis(cdc)
	feemarketFeeGenesis := feemarkettypes.GenesisState{
		Params: feemarkettypes.Params{
			Alpha:                  math.LegacyOneDec(),
			Beta:                   math.LegacyOneDec(),
			Theta:                  math.LegacyOneDec(),
			Delta:                  math.LegacyOneDec(),
			MinBaseFee:             math.LegacyMustNewDecFromStr("0.0025"),
			MinLearningRate:        math.LegacyMustNewDecFromStr("0.5"),
			MaxLearningRate:        math.LegacyMustNewDecFromStr("1.5"),
			TargetBlockUtilization: 1,
			MaxBlockUtilization:    1,
			Window:                 1,
			FeeDenom:               FeeDenom,
			Enabled:                false,
			DistributeFees:         true,
		},
		State: feemarkettypes.State{
			BaseFee:      math.LegacyMustNewDecFromStr("0.0025"),
			LearningRate: math.LegacyOneDec(),
			Window:       []uint64{100},
			Index:        0,
		},
	}
	feemarketFeeGenesisStateBytes, err := json.Marshal(feemarketFeeGenesis)
	if err != nil {
		panic("cannot marshal feemarket genesis state for tests")
	}
	genesisState["feemarket"] = feemarketFeeGenesisStateBytes

	return genesisState
}
