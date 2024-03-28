package cli

import flag "github.com/spf13/pflag"

const (
	FlagMaxAmountOut    = "max-amount-out"
	FlagIncludePoolData = "include-pool-data"
	FlagCalcWithdraw    = "calc-withdraw"
)

func FlagSetMaxAmountOut() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMaxAmountOut, "", "Max amount to be returned from trade")
	return fs
}

func FlagSetIncludePoolData() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(FlagIncludePoolData, false, "Include pool data with response")
	return fs
}

func FlagSetCalcWithdrawableAmount() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(FlagCalcWithdraw, false, "Calculate withdrawable amount")
	return fs
}
