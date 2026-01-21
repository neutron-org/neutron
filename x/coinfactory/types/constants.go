package types

import "cosmossdk.io/store/types"

const ConsensusVersion = 1

// BeforeSendHookGasLimit value is increased to allow existing approved coinfactory hooks to work properly.
// In the next coordinated upgrade, this will become a chain parameter.
var BeforeSendHookGasLimit = types.Gas(500_000)
