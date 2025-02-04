package types

import (
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	vetypes "github.com/skip-mev/slinky/abci/ve/types"
)

// ValidatorParticipation represents a validator's participation in a block creation and oracle prices
// provision.
type ValidatorParticipation struct {
	// ValOperAddress is the validator's operator address.
	ValOperAddress sdktypes.ValAddress
	// BlockVote is the block ID flag the validator voted for.
	BlockVote tmtypes.BlockIDFlag
	// OracleVoteExtension is the validator's vote extension for the oracle module.
	OracleVoteExtension vetypes.OracleVoteExtension
}
