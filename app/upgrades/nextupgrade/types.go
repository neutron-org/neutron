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
