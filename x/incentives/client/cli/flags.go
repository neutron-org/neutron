package cli

import (
	flag "github.com/spf13/pflag"
)

// Flags for incentives module tx commands.
const (
	FlagStartTime = "start-time"
	FlagPerpetual = "perpetual"
	FlagAmount    = "amount"
)

// FlagSetCreateGauge returns flags for creating gauges.
func FlagSetCreateGauge() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStartTime, "", "Timestamp to begin distribution")
	fs.Bool(FlagPerpetual, false, "Perpetual distribution")

	return fs
}
