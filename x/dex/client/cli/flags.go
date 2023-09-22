package cli

import flag "github.com/spf13/pflag"

const (
	FlagMaxAmountOut = "max-amount-out"
)

func FlagSetMaxAmountOut() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMaxAmountOut, "", "Max amount to be returned from trade")
	return fs
}
