package cli

import flag "github.com/spf13/pflag"

const (
	FlagMaxAmountOut    = "max-amount-out"
	FlagIncludePoolData = "include-pool-data"
	FlagCalcWithdraw    = "calc-withdraw"
	FlagPrice           = "price"
	FlagSwapOnDeposit   = "swap-on-deposit"
)

func FlagSetMaxAmountOut() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMaxAmountOut, "", "Max amount to be returned from trade")
	return fs
}

func FlagSetPrice() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPrice, "", "Sell price for limit order")
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

func FlagSetSwapOnDeposit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(FlagSwapOnDeposit, false, "Before BEL swap for deposits")
	return fs
}
