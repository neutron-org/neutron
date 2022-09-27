package types

type MsgSubmitProposal struct {
	Proposals Proposals
	Proposer  string
}

type Proposals struct {
	TextProposal               *TextProposal
	ParamChangeProposal        *ParamChangeProposal
	CommunityPoolSpendProposal *CommunityPoolSpendProposal
}

type TextProposal struct {
	Title       string
	Description string
}

type ParamChangeProposal struct {
	Title       string
	Description string
	Changes     []ParamChange
}

type CommunityPoolSpendProposal struct {
}

type SoftwareUpgradeProposal struct {
}
type CancelSoftwareUpgradeProposal struct {
}

type ClientUpdateProposal struct {
}
type ParamChange struct {
	Subspace string
	Key      string
	Value    string
}
