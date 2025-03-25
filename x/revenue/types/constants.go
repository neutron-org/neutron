package types

import neutronparams "github.com/neutron-org/neutron/v6/app/params"

const (
	ConsensusVersion = 1

	// RewardDenom is the denom used for rewards.
	RewardDenom = neutronparams.DefaultDenom

	// RewardDenomDecimals is the number of decimal places used in the reward denom.
	RewardDenomDecimals = neutronparams.DefaultDenomDecimals
)
