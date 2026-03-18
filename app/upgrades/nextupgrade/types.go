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

type Stack struct {
	elements []string
}

// NewStack creates a new stack instance
func NewStack() *Stack {
	return &Stack{
		elements: make([]string, 0),
	}
}

func (s *Stack) Push(data string) {
	s.elements = append(s.elements, data)
}

// Pop Data from the stack:
func (s *Stack) Pop() string {
	if len(s.elements) == 0 {
		return ""
	}

	data := s.elements[0]
	s.elements[0] = ""
	s.elements = s.elements[1:]
	return data
}

func (s *Stack) Peek() string {
	if len(s.elements) == 0 {
		return ""
	}
	return s.elements[0]
}

func (s *Stack) Size() int {
	return len(s.elements)
}

func (s *Stack) IsEmpty() bool {
	return s.Size() == 0
}
