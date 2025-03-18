package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/dex module sentinel errors

var (
	ErrInvalidTradingPair = sdkerrors.Register(
		ModuleName,
		1102,
		"Invalid token pair:",
	) // "%s<>%s", tokenA, tokenB
	ErrInsufficientShares = sdkerrors.Register(
		ModuleName,
		1104,
		"Insufficient shares:",
	) // "%s does not have %s shares of type %s", address, shares, sharesID
	ErrUnbalancedTxArray = sdkerrors.Register(
		ModuleName,
		1110,
		"Transaction input arrays are not of the same length.",
	)
	ErrValidLimitOrderTrancheNotFound = sdkerrors.Register(
		ModuleName,
		1111,
		"Limit order tranche not found:",
	) // "%d", trancheKey
	ErrCancelEmptyLimitOrder = sdkerrors.Register(
		ModuleName,
		1112,
		"Cannot cancel additional liquidity from limit order tranche:",
	) // "%d", tranche.TrancheKey
	ErrTickOutsideRange = sdkerrors.Register(
		ModuleName,
		1117,
		"abs(tick) + fee must be < 559,680",
	)
	ErrInvalidPoolDenom = sdkerrors.Register(
		ModuleName,
		1118,
		"Denom is not an instance of Neutron PoolDenom",
	)
	ErrInvalidPairIDStr = sdkerrors.Register(
		ModuleName,
		1119,
		"PairID does not conform to pattern TokenA<>TokenB",
	)
	ErrZeroDeposit = sdkerrors.Register(
		ModuleName,
		1120,
		"At least one deposit amount must be > 0.",
	)
	ErrZeroTrueDeposit = sdkerrors.Register(
		ModuleName,
		1121,
		"Cannot deposit single-sided liquidity in tick with opposite liquidity while autoswap is disabled",
	)
	ErrWithdrawEmptyLimitOrder = sdkerrors.Register(
		ModuleName,
		1124,
		"Cannot withdraw additional liqudity from this limit order at this time.",
	)
	ErrZeroSwap = sdkerrors.Register(
		ModuleName,
		1125,
		"MaxAmountIn in must be > 0 for swap.",
	)
	ErrZeroWithdraw = sdkerrors.Register(
		ModuleName,
		1129,
		"Withdraw amount must be > 0.",
	)
	ErrZeroLimitOrder = sdkerrors.Register(
		ModuleName,
		1130,
		"Limit order amount must be > 0.",
	)
	ErrDepositShareUnderflow = sdkerrors.Register(
		ModuleName,
		1133,
		"Deposit amount is too small to issue shares",
	)
	ErrFoKLimitOrderNotFilled = sdkerrors.Register(
		ModuleName,
		1134,
		"Fill Or Kill limit order couldn't be executed in its entirety.",
	)
	ErrInvalidTimeString = sdkerrors.Register(
		ModuleName,
		1135,
		"Time string must be formatted as MM/dd/yyyy HH:mm:ss (ex. 02/05/2023 15:34:56) ",
	)
	ErrGoodTilOrderWithoutExpiration = sdkerrors.Register(
		ModuleName,
		1136,
		"Limit orders of type GOOD_TIL_TIME must supply an ExpirationTime.",
	)
	ErrExpirationOnWrongOrderType = sdkerrors.Register(
		ModuleName,
		1137,
		"Only Limit orders of type GOOD_TIL_TIME can supply an ExpirationTime.",
	)
	ErrInvalidOrderType = sdkerrors.Register(
		ModuleName,
		1138,
		"Order type must be one of: GOOD_TIL_CANCELLED, FILL_OR_KILL, IMMEDIATE_OR_CANCEL, JUST_IN_TIME, or GOOD_TIL_TIME.",
	)
	ErrExpirationTimeInPast = sdkerrors.Register(
		ModuleName,
		1139,
		"Limit order expiration time must be greater than current block time:",
	)
	ErrAllMultiHopRoutesFailed = sdkerrors.Register(
		ModuleName,
		1141,
		"All multihop routes failed limitPrice check or had insufficient liquidity",
	)
	ErrMultihopExitTokensMismatch = sdkerrors.Register(
		ModuleName,
		1142,
		"All multihop routes must have the same exit token",
	)
	ErrMissingMultihopRoute = sdkerrors.Register(
		ModuleName,
		1143,
		"Must supply at least 1 route for multihop swap",
	)
	ErrZeroMaxAmountOut = sdkerrors.Register(
		ModuleName,
		1144,
		"MaxAmountOut must be nil or > 0.",
	)
	ErrInvalidMaxAmountOutForMaker = sdkerrors.Register(
		ModuleName,
		1145,
		"MaxAmountOut can only be set for taker only limit orders.",
	)
	ErrInvalidFee = sdkerrors.Register(
		ModuleName,
		1148,
		"Fee must must a legal fee amount:",
	)
	ErrInvalidAddress = sdkerrors.Register(
		ModuleName,
		1149,
		"Invalid Address",
	)
	ErrRouteWithoutExitToken = sdkerrors.Register(
		ModuleName,
		1150,
		"Each route should specify at least two hops - input and output tokens",
	)
	ErrCycleInHops = sdkerrors.Register(
		ModuleName,
		1151,
		"Hops cannot have cycles",
	)
	ErrZeroExitPrice = sdkerrors.Register(
		ModuleName,
		1152,
		"Cannot have negative or zero exit price",
	)
	ErrDuplicatePoolDeposit = sdkerrors.Register(
		ModuleName,
		1153,
		"Can only provide a single deposit amount for each tick, fee pair",
	)
	ErrLimitPriceNotSatisfied = sdkerrors.Register(
		ModuleName,
		1154,
		"Trade cannot be filled at the specified LimitPrice",
	)
	ErrDexPaused = sdkerrors.Register(
		ModuleName,
		1155,
		"Dex has been paused, all messages are disabled at this time",
	)
	ErrOverJITPerBlockLimit = sdkerrors.Register(
		ModuleName,
		1156,
		"Maximum JIT LimitOrders per block has already been reached",
	)
	ErrInvalidDenom = sdkerrors.Register(
		ModuleName,
		1157,
		"Invalid token denom",
	)
	ErrMultihopEntryTokensMismatch = sdkerrors.Register(
		ModuleName,
		1158,
		"MultihopSwap starting tokens for each route must be the same",
	)
	ErrTradeTooSmall = sdkerrors.Register(
		ModuleName,
		1159,
		"Specified trade will result in a rounded output of 0",
	)
	ErrPriceOutsideRange = sdkerrors.Register(
		ModuleName,
		1160,
		"Invalid price; 0.00000000000000000000000050 < PRICE > 2020125331305056766451886.728",
	)
	ErrInvalidPriceAndTick = sdkerrors.Register(
		ModuleName,
		1161,
		"Only LimitSellPrice or TickIndexInToOut should be specified",
	)
	ErrDepositBehindEnemyLines = sdkerrors.Register(
		ModuleName,
		1162,
		"Cannot deposit at a price below the opposing token's current price",
	)
	ErrCalcTickFromPrice = sdkerrors.Register(
		ModuleName,
		1163,
		"Cannot convert price to int64 tick value",
	)
	ErrNoLiquidity = sdkerrors.Register(
		ModuleName,
		1164,
		"No tradable liquidity below LimitSellPrice",
	)
	ErrZeroMinAverageSellPrice = sdkerrors.Register(
		ModuleName,
		1165,
		"MinAverageSellPrice must be nil or > 0",
	)
	ErrDoubleSidedSwapOnDeposit = sdkerrors.Register(
		ModuleName,
		1166,
		"Swap on deposit cannot be performed for Token0 and Token1",
	)
	ErrSwapOnDepositWithoutAutoswap = sdkerrors.Register(
		ModuleName,
		1167,
		"Cannot disable autoswap when using swap_on_deposit",
	)
	ErrSwapOnDepositSlopToleranceNotSatisfied = sdkerrors.Register(
		ModuleName,
		1168,
		"Swap on deposit true price is less than minimum allowed price",
	)
	ErrInvalidSlopTolerance = sdkerrors.Register(
		ModuleName,
		1169,
		"Slop tolerance must be between 0 and 10000",
	)
)
