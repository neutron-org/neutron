package nextupgrade

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type IBCRateLimitsExecuteMessage struct {
	GrantRole  *GrantRole  `json:"grant_role,omitempty"`
	RevokeRole *RevokeRole `json:"revoke_role,omitempty"`
}

type GrantRole struct {
	Signer string   `json:"signer"`
	Roles  []string `json:"roles"`
}

type RevokeRole struct {
	Signer string   `json:"signer"`
	Roles  []string `json:"roles"`
}

// Redelegation represents redelegation plan for a single new validator.
type Redelegation struct {
	// ValidatorAddress is the address of the new validator to which the delegation is being redelegated.
	ValidatorAddress sdk.ValAddress
	// RedelegationMsgs is a list of MsgBeginRedelegate messages that represent the redelegation plan.
	// Possible to have many-to-one redelegation to the same validator.
	RedelegationMsgs []stakingtypes.MsgBeginRedelegate
	// Redelegated is the amount of shares to be redelegated to the new validator.
	Redelegated math.LegacyDec
}
