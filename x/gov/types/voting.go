package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// InstantiateMsg corresponding to the Rust type `neutron-voting::msg::InstantiateMsg`
type InstantiateMsg struct {
	Owner string `json:"owner"`
}

// QueryMsg corresponding to the Rust enum `neutron-voting::msg::QueryMsg`
//
// NOTE: For covenience, we don't include other enum variants, as they are not needed here
type QueryMsg struct {
	VotingPower  *VotingPowerQuery  `json:"voting_power,omitempty"`
	VotingPowers *VotingPowersQuery `json:"voting_powers,omitempty"`
}

// VotingPowerQuery corresponding to the Rust enum variant `neutron-voting::msg::QueryMsg::VotingPower`
type VotingPowerQuery struct {
	User string `json:"user,omitempty"`
}

// VotingPowersQuery corresponding to the Rust enum variant `neutron-voting::msg::QueryMsg::VotingPowers`
type VotingPowersQuery struct{}

// VotingPowerResponse corresponding to the `voting_powers` query's respons type's repeating element
type VotingPowerResponse struct {
	User        string   `json:"user"`
	VotingPower sdk.Uint `json:"voting_power"`
}

// VotingPowersResponse corresponding to the response type of the `voting_powers` query
type VotingPowersResponse []VotingPowerResponse
