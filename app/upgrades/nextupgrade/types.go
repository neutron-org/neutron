package nextupgrade

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

type VotingRegistryExecuteMessage struct {
	DeactivateVotingVault *DeactivateVotingVault `json:"deactivate_voting_vault,omitempty"`
	UpdateConfig          *UpdateConfig          `json:"update_config,omitempty"`
}

type DeactivateVotingVault struct {
	VotingVaultContract string `json:"voting_vault_contract"`
}

type UpdateConfig struct {
	Owner string `json:"owner"`
}

type VotingVault struct {
	Address string `json:"address"`
}
