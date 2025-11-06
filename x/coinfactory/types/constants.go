package types

const ConsensusVersion = 1

// TrackBeforeSendGasLimit value is increased to allow existing approved coinfactory hooks to work properly.
// In the next coordinated upgrade, this will become a chain parameter.
var TrackBeforeSendGasLimit = uint64(500_000)
