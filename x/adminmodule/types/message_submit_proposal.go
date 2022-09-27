package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramChange "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (m *MsgSubmitProposal) GetProposal() govtypes.Content {
	if m.Proposals.TextProposal != nil {
		textProposal := govtypes.TextProposal{
			Title:       m.Proposals.TextProposal.Title,
			Description: m.Proposals.TextProposal.Description,
		}
		return &textProposal
	}

	if m.Proposals.TextProposal != nil {
		prop := m.Proposals.ParamChangeProposal
		var paramsChange []paramChange.ParamChange
		for _, param := range prop.Changes {
			parameterChange := paramChange.ParamChange{
				Subspace: param.Subspace,
				Key:      param.Key,
				Value:    param.Value,
			}
			paramsChange = append(paramsChange, parameterChange)
		}
		paramsProposal := paramChange.ParameterChangeProposal{
			Title:       prop.Title,
			Description: prop.Description,
			Changes:     paramsChange,
		}
		return &paramsProposal
	}

	return nil
}

func (m *MsgSubmitProposal) Route() string {
	return RouterKey
}

func (m *MsgSubmitProposal) Type() string {
	return "SubmitProposal"
}

func (m *MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(m.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
}

// ValidateBasic implements Msg
func (m *MsgSubmitProposal) ValidateBasic() error {
	if m.Proposer == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, m.Proposer)
	}

	content := m.GetProposal()
	if content == nil {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "missing content")
	}
	if !govtypes.IsValidProposalType(content.ProposalType()) {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalType, content.ProposalType())
	}
	if err := content.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
