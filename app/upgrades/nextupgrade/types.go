package nextupgrade

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

type Stack struct {
	elements []sdk.ValAddress
}

// NewStack creates a new stack instance
func NewStack() *Stack {
	return &Stack{
		elements: make([]sdk.ValAddress, 0),
	}
}

func (s *Stack) Push(data sdk.ValAddress) {
	s.elements = append(s.elements, data)
}

// Pop Data from the stack:
func (s *Stack) Pop() sdk.ValAddress {
	if len(s.elements) == 0 {
		return nil
	}

	data := s.elements[0]
	s.elements[0] = nil
	s.elements = s.elements[1:]
	return data
}

func (s *Stack) Peek() sdk.ValAddress {
	if len(s.elements) == 0 {
		return nil
	}
	return s.elements[0]
}

func (s *Stack) Size() int {
	return len(s.elements)
}

func (s *Stack) IsEmpty() bool {
	return s.Size() == 0
}
