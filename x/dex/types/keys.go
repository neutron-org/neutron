package types

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/utils"
)

const (
	// ModuleName defines the module name
	ModuleName = "dex"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_dex"
)

const (
	// TickLiquidityKeyPrefix is the prefix to retrieve all TickLiquidity
	TickLiquidityKeyPrefix = "TickLiquidity/value/"

	// LimitOrderTrancheUserKeyPrefix is the prefix to retrieve all LimitOrderTrancheUser
	LimitOrderTrancheUserKeyPrefix = "LimitOrderTrancheUser/value"

	// InactiveLimitOrderTrancheKeyPrefix is the prefix to retrieve all InactiveLimitOrderTranche
	InactiveLimitOrderTrancheKeyPrefix = "InactiveLimitOrderTranche/value/"

	// LimitOrderExpirationKeyPrefix is the prefix to retrieve all LimitOrderExpiration
	LimitOrderExpirationKeyPrefix = "LimitOrderExpiration/value/"

	// PoolIDKeyPrefix is the prefix to retrieve all PoolIds or retrieve a specific pool by pair+tick+fee
	PoolIDKeyPrefix = "Pool/id/"

	// PoolMetadataKeyPrefix is the prefix to retrieve all PoolMetadata
	PoolMetadataKeyPrefix = "PoolMetadata/value/"

	// PoolCountKeyPrefix is the prefix to retrieve the Pool count
	PoolCountKeyPrefix = "Pool/count/"

	// ParamsKey is the prefix to retrieve params
	ParamsKey = "Params/value/"
)

func KeyPrefix(p string) []byte {
	key := []byte(p)
	key = append(key, []byte("/")...)
	return key
}

func TickIndexToBytes(tickTakerToMaker int64) []byte {
	key := make([]byte, 9)
	if tickTakerToMaker < 0 {
		copy(key[1:], sdk.Uint64ToBigEndian(uint64(tickTakerToMaker)))
	} else {
		copy(key[:1], []byte{0x01})
		copy(key[1:], sdk.Uint64ToBigEndian(uint64(tickTakerToMaker)))
	}

	return key
}

func BytesToTickIndex(bz []byte) (int64, error) {
	if len(bz) != 9 {
		return 0, errors.New("input should be 9 bytes long")
	}

	isNegative := bz[0] == 0
	tickTakerToMaker := sdk.BigEndianToUint64(bz[1:])

	if isNegative {
		return int64(-tickTakerToMaker), nil
	}
	// else
	return int64(tickTakerToMaker), nil
}

// LimitOrderTrancheUserKey returns the store key to retrieve a LimitOrderTrancheUser from the index fields
func LimitOrderTrancheUserKey(address, trancheKey string) []byte {
	var key []byte

	addressBytes := []byte(address)
	key = append(key, addressBytes...)
	key = append(key, []byte("/")...)

	trancheKeyBytes := []byte(trancheKey)
	key = append(key, trancheKeyBytes...)
	key = append(key, []byte("/")...)

	return key
}

func LimitOrderTrancheUserAddressPrefix(address string) []byte {
	key := KeyPrefix(LimitOrderTrancheUserKeyPrefix)
	addressBytes := []byte(address)
	key = append(key, addressBytes...)
	key = append(key, []byte("/")...)

	return key
}

func TimeBytes(timestamp time.Time) []byte {
	var unixSecs uint64
	// If timestamp is 0 use that instead of returning long negative number for unix time
	if !timestamp.IsZero() {
		unixSecs = uint64(timestamp.Unix())
	}

	str := utils.Uint64ToSortableString(unixSecs)
	return []byte(str)
}

func TickLiquidityLimitOrderPrefix(
	tradePairID *TradePairID,
	tickIndexTakerTomMaker int64,
) []byte {
	key := KeyPrefix(TickLiquidityKeyPrefix)

	pairIDBytes := []byte(tradePairID.MustPairID().CanonicalString())
	key = append(key, pairIDBytes...)
	key = append(key, []byte("/")...)

	makerDenomBytes := []byte(tradePairID.MakerDenom)
	key = append(key, makerDenomBytes...)
	key = append(key, []byte("/")...)

	tickIndexBytes := TickIndexToBytes(tickIndexTakerTomMaker)
	key = append(key, tickIndexBytes...)
	key = append(key, []byte("/")...)

	liquidityTypeBytes := []byte(LiquidityTypeLimitOrder)
	key = append(key, liquidityTypeBytes...)
	key = append(key, []byte("/")...)

	return key
}

func TickLiquidityPrefix(tradePairID *TradePairID) []byte {
	var key []byte
	key = append(KeyPrefix(TickLiquidityKeyPrefix), KeyPrefix(tradePairID.MustPairID().CanonicalString())...)
	key = append(key, KeyPrefix(tradePairID.MakerDenom)...)

	return key
}

func LimitOrderExpirationKey(
	goodTilDate time.Time,
	trancheRef []byte,
) []byte {
	var key []byte

	goodTilDateBytes := TimeBytes(goodTilDate)
	key = append(key, goodTilDateBytes...)
	key = append(key, []byte("/")...)

	key = append(key, trancheRef...)
	key = append(key, []byte("/")...)

	return key
}

func PoolIDKey(
	pairID *PairID,
	tickIndex int64,
	fee uint64,
) []byte {
	key := []byte(pairID.CanonicalString())
	key = append(key, []byte("/")...)

	tickIndexBytes := TickIndexToBytes(tickIndex)
	key = append(key, tickIndexBytes...)
	key = append(key, []byte("/")...)

	feeBytes := sdk.Uint64ToBigEndian(fee)
	key = append(key, feeBytes...)
	key = append(key, []byte("/")...)

	return key
}

// Deposit Event Attributes
const (
	DepositEventKey                = "DepositLP"
	DepositEventCreator            = "Creator"
	DepositEventToken0             = "TokenZero"
	DepositEventToken1             = "TokenOne"
	DepositEventPrice              = "TickIndex"
	DepositEventFee                = "Fee"
	DepositEventReceiver           = "Receiver"
	DepositEventReserves0Deposited = "ReservesZeroDeposited"
	DepositEventReserves1Deposited = "ReservesOneDeposited"
	DepositEventSharesMinted       = "SharesMinted"
)

// Withdraw Event Attributes
const (
	WithdrawEventKey                = "WithdrawLP"
	WithdrawEventCreator            = "Creator"
	WithdrawEventToken0             = "TokenZero"
	WithdrawEventToken1             = "TokenOne"
	WithdrawEventPrice              = "TickIndex"
	WithdrawEventFee                = "Fee"
	WithdrawEventReceiver           = "Receiver"
	WithdrawEventReserves0Withdrawn = "ReservesZeroWithdrawn"
	WithdrawEventReserves1Withdrawn = "ReservesOneWithdrawn"
	WithdrawEventSharesRemoved      = "SharesRemoved"
)

// Multihop-Swap Event Attributes
const (
	MultihopSwapEventKey       = "MultihopSwap"
	MultihopSwapEventCreator   = "Creator"
	MultihopSwapEventReceiver  = "Receiver"
	MultihopSwapEventTokenIn   = "TokenIn"
	MultihopSwapEventTokenOut  = "TokenOut"
	MultihopSwapEventAmountIn  = "AmountIn"
	MultihopSwapEventAmountOut = "AmountOut"
	MultihopSwapEventRoute     = "Route"
	MultihopSwapEventDust      = "Dust"
)

// Place LimitOrder Event Attributes
const (
	PlaceLimitOrderEventKey        = "PlaceLimitOrder"
	PlaceLimitOrderEventCreator    = "Creator"
	PlaceLimitOrderEventReceiver   = "Receiver"
	PlaceLimitOrderEventToken0     = "TokenZero"
	PlaceLimitOrderEventToken1     = "TokenOne"
	PlaceLimitOrderEventTokenIn    = "TokenIn"
	PlaceLimitOrderEventTokenOut   = "TokenOut"
	PlaceLimitOrderEventAmountIn   = "AmountIn"
	PlaceLimitOrderEventLimitTick  = "LimitTick"
	PlaceLimitOrderEventOrderType  = "OrderType"
	PlaceLimitOrderEventShares     = "Shares"
	PlaceLimitOrderEventTrancheKey = "TrancheKey"
)

// Withdraw LimitOrder Event Attributes
const (
	WithdrawFilledLimitOrderEventKey        = "WithdrawLimitOrder"
	WithdrawFilledLimitOrderEventCreator    = "Creator"
	WithdrawFilledLimitOrderEventToken0     = "TokenZero"
	WithdrawFilledLimitOrderEventToken1     = "TokenOne"
	WithdrawFilledLimitOrderEventTokenIn    = "TokenIn"
	WithdrawFilledLimitOrderEventTokenOut   = "TokenOut"
	WithdrawFilledLimitOrderEventTrancheKey = "TrancheKey"
	WithdrawFilledLimitOrderEventAmountOut  = "AmountOut"
)

// Cancel LimitOrder Event Attributes
const (
	CancelLimitOrderEventKey        = "CancelLimitOrder"
	CancelLimitOrderEventCreator    = "Creator"
	CancelLimitOrderEventToken0     = "TokenZero"
	CancelLimitOrderEventToken1     = "TokenOne"
	CancelLimitOrderEventTokenIn    = "TokenIn"
	CancelLimitOrderEventTokenOut   = "TokenOut"
	CancelLimitOrderEventTrancheKey = "TrancheKey"
	CancelLimitOrderEventAmountOut  = "AmountOut"
)

// Tick Update Event Attributes
const (
	EventTypeTickUpdate       = "TickUpdate"
	TickUpdateEventKey        = "TickUpdate"
	TickUpdateEventToken0     = "TokenZero"
	TickUpdateEventToken1     = "TokenOne"
	TickUpdateEventTokenIn    = "TokenIn"
	TickUpdateEventTickIndex  = "TickIndex"
	TickUpdateEventFee        = "Fee"
	TickUpdateEventTrancheKey = "TrancheKey"
	TickUpdateEventReserves   = "Reserves"
)

const (
	EventTypeGoodTilPurgeHitGasLimit = "GoodTilPurgeHitGasLimit"
	GoodTilPurgeHitGasLimitEventGas  = "Gas"
)

const (
	// NOTE: have to add letter so that LP deposits are indexed ahead of LimitOrders
	LiquidityTypePoolReserves = "A_PoolDeposit"
	LiquidityTypeLimitOrder   = "B_LODeposit"
)

func JITGoodTilTime() time.Time {
	return time.Time{}
}

const (
	// NOTE: This number is based current cost of all operations in EndBlock,
	// if that changes this value must be updated to ensure there is enough
	// remaining gas (weak proxy for timeoutPrepareProposal) to complete endBlock
	GoodTilPurgeGasBuffer = 50_000
	ExpiringLimitOrderGas = 10_000
)
