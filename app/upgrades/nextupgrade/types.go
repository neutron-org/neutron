package nextupgrade

type IBCRateLimitsExecuteMessage struct {
	GrantRole  *GrantRole  `json:"grant_role"`
	RevokeRole *RevokeRole `json:"revoke_role"`
}

type GrantRole struct {
	Signer string   `json:"signer"`
	Roles  []string `json:"role"`
}

type RevokeRole struct {
	Signer string   `json:"signer"`
	Roles  []string `json:"role"`
}
