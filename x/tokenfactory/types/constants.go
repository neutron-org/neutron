package types

const ConsensusVersion = 2

// SudoHookGasLimit value is increased to allow existing approved tokenfactory hooks to work properly.
// In the next coordinated upgrade, this will become a chain parameter.
var SudoHookGasLimit = uint64(500_000)
