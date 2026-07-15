package math

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/math"
)

// NOTE: This file is nearly direct copy from cosmossdk.io/math/dec.go @v1.01
// The precision has been changed from 18 to 27

// NOTE: never use new(Dec) or else we will panic unmarshalling into the
// nil embedded big.Int
type PrecDec struct {
	i *big.Int
}

const (
	// number of decimal places
	Precision = 27

	// bits required to represent the above precision
	// Ceiling[Log2[10^Precision - 1]]
	PrecDecimalPrecisionBits = 90

	// decimalTruncateBits is the minimum number of bits removed
	// by a truncate operation. It is equal to
	// Floor[Log2[10^Precision - 1]].
	decimalTruncateBits = PrecDecimalPrecisionBits - 1

	MaxBitLen = 256

	maxPrecDecBitLen = MaxBitLen + decimalTruncateBits

	// max number of iterations in ApproxRoot function
	maxApproxRootIterations = 300
)

var (
	precisionReuse       = new(big.Int).Exp(big.NewInt(10), big.NewInt(Precision), nil)
	fivePrecision        = new(big.Int).Quo(precisionReuse, big.NewInt(2))
	precisionMultipliers []*big.Int
	zeroInt              = big.NewInt(0)
	oneInt               = big.NewInt(1)
	tenInt               = big.NewInt(10)
)

// PrecDecimal errors
var (
	ErrEmptyPrecDecimalStr      = errors.New("decimal string cannot be empty")
	ErrInvalidPrecDecimalLength = errors.New("invalid decimal length")
	ErrInvalidPrecDecimalStr    = errors.New("invalid decimal string")
)

// Set precision multipliers
func init() {
	precisionMultipliers = make([]*big.Int, Precision+1)
	for i := 0; i <= Precision; i++ {
		precisionMultipliers[i] = calcPrecisionMultiplier(int64(i))
	}
}

func precisionInt() *big.Int {
	return new(big.Int).Set(precisionReuse)
}

func ZeroPrecDec() PrecDec     { return PrecDec{new(big.Int).Set(zeroInt)} }
func OnePrecDec() PrecDec      { return PrecDec{precisionInt()} }
func SmallestPrecDec() PrecDec { return PrecDec{new(big.Int).Set(oneInt)} }

// calculate the precision multiplier
func calcPrecisionMultiplier(prec int64) *big.Int {
	if prec < 0 {
		panic(fmt.Sprintf("negative precision %v", prec))
	}

	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	zerosToAdd := Precision - prec
	multiplier := new(big.Int).Exp(tenInt, big.NewInt(zerosToAdd), nil)
	return multiplier
}

// get the precision multiplier, do not mutate result
func precisionMultiplier(prec int64) *big.Int {
	if prec < 0 {
		panic(fmt.Sprintf("negative precision %v", prec))
	}

	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	return precisionMultipliers[prec]
}

// create a new PrecDec from integer assuming whole number
func NewPrecDec(i int64) PrecDec {
	return NewPrecDecWithPrec(i, 0)
}

// create a new PrecDec from integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewPrecDecWithPrec(i, prec int64) PrecDec {
	return PrecDec{
		new(big.Int).Mul(big.NewInt(i), precisionMultiplier(prec)),
	}
}

// create a new PrecDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewPrecDecFromBigInt(i *big.Int) PrecDec {
	return NewPrecDecFromBigIntWithPrec(i, 0)
}

// create a new PrecDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewPrecDecFromBigIntWithPrec(i *big.Int, prec int64) PrecDec {
	return PrecDec{
		new(big.Int).Mul(i, precisionMultiplier(prec)),
	}
}

// create a new PrecDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewPrecDecFromInt(i math.Int) PrecDec {
	return NewPrecDecFromIntWithPrec(i, 0)
}

// create a new PrecDec from big integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewPrecDecFromIntWithPrec(i math.Int, prec int64) PrecDec {
	return PrecDec{
		new(big.Int).Mul(i.BigInt(), precisionMultiplier(prec)),
	}
}

// create a decimal from an input decimal string.
// valid must come in the form:
//
//	(-) whole integers (.) decimal integers
//
// examples of acceptable input include:
//
//	-123.456
//	456.7890
//	345
//	-456789
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
func NewPrecDecFromStr(str string) (PrecDec, error) {
	// first extract any negative symbol
	neg := false
	if len(str) > 0 && str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return PrecDec{}, ErrEmptyPrecDecimalStr
	}

	strs := strings.Split(str, ".")
	lenPrecDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenPrecDecs = len(strs[1])
		if lenPrecDecs == 0 || len(combinedStr) == 0 {
			return PrecDec{}, ErrInvalidPrecDecimalLength
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return PrecDec{}, ErrInvalidPrecDecimalStr
	}

	if lenPrecDecs > Precision {
		return PrecDec{}, fmt.Errorf("value '%s' exceeds max precision by %d decimal places: max precision %d", str, Precision-lenPrecDecs, Precision)
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := Precision - lenPrecDecs
	zeros := strings.Repeat("0", zerosToAdd)
	combinedStr += zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return PrecDec{}, fmt.Errorf("failed to set decimal string with base 10: %s", combinedStr)
	}
	if combined.BitLen() > maxPrecDecBitLen {
		return PrecDec{}, fmt.Errorf("decimal '%s' out of range; bitLen: got %d, max %d", str, combined.BitLen(), maxPrecDecBitLen)
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}

	return PrecDec{combined}, nil
}

// PrecDecimal from string, panic on error
func MustNewPrecDecFromStr(s string) PrecDec {
	dec, err := NewPrecDecFromStr(s)
	if err != nil {
		panic(err)
	}
	return dec
}

func (d PrecDec) IsNil() bool            { return d.i == nil }                     // is decimal nil
func (d PrecDec) IsZero() bool           { return (d.i).Sign() == 0 }              // is equal to zero
func (d PrecDec) IsNegative() bool       { return (d.i).Sign() == -1 }             // is negative
func (d PrecDec) IsPositive() bool       { return (d.i).Sign() == 1 }              // is positive
func (d PrecDec) Equal(d2 PrecDec) bool  { return (d.i).Cmp(d2.i) == 0 }           // equal decimals
func (d PrecDec) GT(d2 PrecDec) bool     { return (d.i).Cmp(d2.i) > 0 }            // greater than
func (d PrecDec) GTE(d2 PrecDec) bool    { return (d.i).Cmp(d2.i) >= 0 }           // greater than or equal
func (d PrecDec) LT(d2 PrecDec) bool     { return (d.i).Cmp(d2.i) < 0 }            // less than
func (d PrecDec) LTE(d2 PrecDec) bool    { return (d.i).Cmp(d2.i) <= 0 }           // less than or equal
func (d PrecDec) Neg() PrecDec           { return PrecDec{new(big.Int).Neg(d.i)} } // reverse the decimal sign
func (d PrecDec) NegMut() PrecDec        { d.i.Neg(d.i); return d }                // reverse the decimal sign, mutable
func (d PrecDec) Abs() PrecDec           { return PrecDec{new(big.Int).Abs(d.i)} } // absolute value
func (d PrecDec) AbsMut() PrecDec        { d.i.Abs(d.i); return d }                // absolute value, mutable
func (d PrecDec) Set(d2 PrecDec) PrecDec { d.i.Set(d2.i); return d }               // set to existing dec value
func (d PrecDec) Clone() PrecDec         { return PrecDec{new(big.Int).Set(d.i)} } // clone new dec

// BigInt returns a copy of the underlying big.Int.
func (d PrecDec) BigInt() *big.Int {
	if d.IsNil() {
		return nil
	}

	cp := new(big.Int)
	return cp.Set(d.i)
}

func (d PrecDec) ImmutOp(op func(PrecDec, PrecDec) PrecDec, d2 PrecDec) PrecDec {
	return op(d.Clone(), d2)
}

func (d PrecDec) ImmutOpInt(op func(PrecDec, math.Int) PrecDec, d2 math.Int) PrecDec {
	return op(d.Clone(), d2)
}

func (d PrecDec) ImmutOpInt64(op func(PrecDec, int64) PrecDec, d2 int64) PrecDec {
	// TODO: use already allocated operand bigint to avoid
	// newint each time, add mutex for race condition
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/11166
	return op(d.Clone(), d2)
}

func (d PrecDec) SetInt64(i int64) PrecDec {
	d.i.SetInt64(i)
	d.i.Mul(d.i, precisionReuse)
	return d
}

// addition
func (d PrecDec) Add(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.AddMut, d2)
}

// mutable addition
func (d PrecDec) AddMut(d2 PrecDec) PrecDec {
	d.i.Add(d.i, d2.i)

	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// subtraction
func (d PrecDec) Sub(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.SubMut, d2)
}

// mutable subtraction
func (d PrecDec) SubMut(d2 PrecDec) PrecDec {
	d.i.Sub(d.i, d2.i)

	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// multiplication
func (d PrecDec) Mul(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.MulMut, d2)
}

// mutable multiplication
func (d PrecDec) MulMut(d2 PrecDec) PrecDec {
	d.i.Mul(d.i, d2.i)
	chopped := chopPrecisionAndRound(d.i)

	if chopped.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	*d.i = *chopped
	return d
}

// multiplication truncate
func (d PrecDec) MulTruncate(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.MulTruncateMut, d2)
}

// mutable multiplication truncage
func (d PrecDec) MulTruncateMut(d2 PrecDec) PrecDec {
	d.i.Mul(d.i, d2.i)
	chopPrecisionAndTruncate(d.i)

	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// multiplication
func (d PrecDec) MulInt(i math.Int) PrecDec {
	return d.ImmutOpInt(PrecDec.MulIntMut, i)
}

func (d PrecDec) MulIntMut(i math.Int) PrecDec {
	d.i.Mul(d.i, i.BigInt())
	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// MulInt64 - multiplication with int64
func (d PrecDec) MulInt64(i int64) PrecDec {
	return d.ImmutOpInt64(PrecDec.MulInt64Mut, i)
}

func (d PrecDec) MulInt64Mut(i int64) PrecDec {
	d.i.Mul(d.i, big.NewInt(i))

	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// quotient
func (d PrecDec) Quo(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.QuoMut, d2)
}

var squaredPrecisionReuse = new(big.Int).Mul(precisionReuse, precisionReuse)

// mutable quotient
func (d PrecDec) QuoMut(d2 PrecDec) PrecDec {
	// multiply by precision twice
	d.i.Mul(d.i, squaredPrecisionReuse)
	d.i.Quo(d.i, d2.i)

	chopPrecisionAndRound(d.i)
	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// quotient truncate
func (d PrecDec) QuoTruncate(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.QuoTruncateMut, d2)
}

// mutable quotient truncate
func (d PrecDec) QuoTruncateMut(d2 PrecDec) PrecDec {
	// multiply precision twice
	d.i.Mul(d.i, squaredPrecisionReuse)
	d.i.Quo(d.i, d2.i)

	chopPrecisionAndTruncate(d.i)
	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// quotient, round up
func (d PrecDec) QuoRoundUp(d2 PrecDec) PrecDec {
	return d.ImmutOp(PrecDec.QuoRoundupMut, d2)
}

// mutable quotient, round up
func (d PrecDec) QuoRoundupMut(d2 PrecDec) PrecDec {
	// multiply precision twice
	d.i.Mul(d.i, squaredPrecisionReuse)
	d.i.Quo(d.i, d2.i)

	chopPrecisionAndRoundUp(d.i)
	if d.i.BitLen() > maxPrecDecBitLen {
		panic("Int overflow")
	}
	return d
}

// quotient
func (d PrecDec) QuoInt(i math.Int) PrecDec {
	return d.ImmutOpInt(PrecDec.QuoIntMut, i)
}

func (d PrecDec) QuoIntMut(i math.Int) PrecDec {
	d.i.Quo(d.i, i.BigInt())
	return d
}

// QuoInt64 - quotient with int64
func (d PrecDec) QuoInt64(i int64) PrecDec {
	return d.ImmutOpInt64(PrecDec.QuoInt64Mut, i)
}

func (d PrecDec) QuoInt64Mut(i int64) PrecDec {
	d.i.Quo(d.i, big.NewInt(i))
	return d
}

// ApproxRoot returns an approximate estimation of a PrecDec's positive real nth root
// using Newton's method (where n is positive). The algorithm starts with some guess and
// computes the sequence of improved guesses until an answer converges to an
// approximate answer.  It returns `|d|.ApproxRoot() * -1` if input is negative.
// A maximum number of 100 iterations is used a backup boundary condition for
// cases where the answer never converges enough to satisfy the main condition.
func (d PrecDec) ApproxRoot(root uint64) (guess PrecDec, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = errors.New("out of bounds")
			}
		}
	}()

	if d.IsNegative() {
		absRoot, err := d.Neg().ApproxRoot(root)
		return absRoot.NegMut(), err
	}

	// One decimal, that we invalidate later. Helps us save a heap allocation.
	scratchOnePrecDec := OnePrecDec()
	if root == 1 || d.IsZero() || d.Equal(scratchOnePrecDec) {
		return d, nil
	}

	if root == 0 {
		return scratchOnePrecDec, nil
	}

	guess, delta := scratchOnePrecDec, OnePrecDec()
	smallestPrecDec := SmallestPrecDec()

	for iter := 0; delta.AbsMut().GT(smallestPrecDec) && iter < maxApproxRootIterations; iter++ {
		// Set prev = guess^{root - 1}, with an optimization for sqrt
		// where root=2 => prev = guess. (And thus no extra heap allocations)
		prev := guess
		if root != 2 {
			prev = guess.Power(root - 1)
		}
		if prev.IsZero() {
			prev = smallestPrecDec
		}
		delta.Set(d).QuoMut(prev)
		delta.SubMut(guess)
		// delta = delta / root.
		// We optimize for sqrt, where root=2 => delta = delta >> 1
		if root == 2 {
			delta.i.Rsh(delta.i, 1)
		} else {
			delta.QuoInt64Mut(int64(root))
		}

		guess.AddMut(delta)
	}

	return guess, nil
}

// Power returns the result of raising to a positive integer power
func (d PrecDec) Power(power uint64) PrecDec {
	res := PrecDec{new(big.Int).Set(d.i)}
	return res.PowerMut(power)
}

func (d PrecDec) PowerMut(power uint64) PrecDec {
	if power == 0 {
		d.SetInt64(1)
		return d
	}
	tmp := OnePrecDec()

	for i := power; i > 1; {
		if i%2 != 0 {
			tmp.MulMut(d)
		}
		i /= 2
		d.MulMut(d)
	}

	return d.MulMut(tmp)
}

// ApproxSqrt is a wrapper around ApproxRoot for the common special case
// of finding the square root of a number. It returns -(sqrt(abs(d)) if input is negative.
func (d PrecDec) ApproxSqrt() (PrecDec, error) {
	return d.ApproxRoot(2)
}

// is integer, e.g. decimals are zero
func (d PrecDec) IsInteger() bool {
	return new(big.Int).Rem(d.i, precisionReuse).Sign() == 0
}

// Format decimal state
func (d PrecDec) Format(s fmt.State) {
	_, err := s.Write([]byte(d.String()))
	if err != nil {
		panic(err)
	}
}

func (d PrecDec) String() string {
	if d.i == nil {
		return d.i.String()
	}

	isNeg := d.IsNegative()

	if isNeg {
		d = d.Neg()
	}

	bzInt, err := d.i.MarshalText()
	if err != nil {
		return ""
	}
	inputSize := len(bzInt)

	var bzStr []byte

	// TODO: Remove trailing zeros
	// case 1, purely decimal
	if inputSize <= Precision {
		bzStr = make([]byte, Precision+2)

		// 0. prefix
		bzStr[0] = byte('0')
		bzStr[1] = byte('.')

		// set relevant digits to 0
		for i := 0; i < Precision-inputSize; i++ {
			bzStr[i+2] = byte('0')
		}

		// set final digits
		copy(bzStr[2+(Precision-inputSize):], bzInt)
	} else {
		// inputSize + 1 to account for the decimal point that is being added
		bzStr = make([]byte, inputSize+1)
		decPointPlace := inputSize - Precision

		copy(bzStr, bzInt[:decPointPlace])                   // pre-decimal digits
		bzStr[decPointPlace] = byte('.')                     // decimal point
		copy(bzStr[decPointPlace+1:], bzInt[decPointPlace:]) // post-decimal digits
	}

	if isNeg {
		return "-" + string(bzStr)
	}

	return string(bzStr)
}

// Float64 returns the float64 representation of a PrecDec.
// Will return the error if the conversion failed.
func (d PrecDec) Float64() (float64, error) {
	return strconv.ParseFloat(d.String(), 64)
}

// MustFloat64 returns the float64 representation of a PrecDec.
// Would panic if the conversion failed.
func (d PrecDec) MustFloat64() float64 {
	value, err := strconv.ParseFloat(d.String(), 64)
	if err != nil {
		panic(err)
	}

	return value
}

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
//
// Mutates the input. Use the non-mutative version if that is undesired
func chopPrecisionAndRound(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(fivePrecision) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}

func chopPrecisionAndRoundUp(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		// truncate since d is negative...
		chopPrecisionAndTruncate(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	return quo.Add(quo, oneInt)
}

func chopPrecisionAndRoundNonMutative(d *big.Int) *big.Int {
	tmp := new(big.Int).Set(d)
	return chopPrecisionAndRound(tmp)
}

// RoundInt64 rounds the decimal using bankers rounding
func (d PrecDec) RoundInt64() int64 {
	chopped := chopPrecisionAndRoundNonMutative(d.i)
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// RoundInt round the decimal using bankers rounding
func (d PrecDec) RoundInt() math.Int {
	return math.NewIntFromBigInt(chopPrecisionAndRoundNonMutative(d.i))
}

// chopPrecisionAndTruncate is similar to chopPrecisionAndRound,
// but always rounds down. It does not mutate the input.
func chopPrecisionAndTruncate(d *big.Int) {
	d.Quo(d, precisionReuse)
}

func chopPrecisionAndTruncateNonMutative(d *big.Int) *big.Int {
	tmp := new(big.Int).Set(d)
	chopPrecisionAndTruncate(tmp)
	return tmp
}

// TruncateInt64 truncates the decimals from the number and returns an int64
func (d PrecDec) TruncateInt64() int64 {
	chopped := chopPrecisionAndTruncateNonMutative(d.i)
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// TruncateInt truncates the decimals from the number and returns an Int
func (d PrecDec) TruncateInt() math.Int {
	return math.NewIntFromBigInt(chopPrecisionAndTruncateNonMutative(d.i))
}

// TruncatePrecDec truncates the decimals from the number and returns a PrecDec
func (d PrecDec) TruncatePrecDec() PrecDec {
	return NewPrecDecFromBigInt(chopPrecisionAndTruncateNonMutative(d.i))
}

// Ceil returns the smallest integer value (as a decimal) that is greater than
// or equal to the given decimal.
func (d PrecDec) Ceil() PrecDec {
	tmp := new(big.Int).Set(d.i)

	quo, rem := tmp, big.NewInt(0)
	quo, rem = quo.QuoRem(tmp, precisionReuse, rem)

	// no need to round with a zero remainder regardless of sign
	if rem.Cmp(zeroInt) == 0 {
		return NewPrecDecFromBigInt(quo)
	}

	if rem.Sign() == -1 {
		return NewPrecDecFromBigInt(quo)
	}

	return NewPrecDecFromBigInt(quo.Add(quo, oneInt))
}

// MaxSortablePrecDec is the largest PrecDec that can be passed into SortablePrecDecBytes()
// Its negative form is the least PrecDec that can be passed in.
var MaxSortablePrecDec PrecDec

func init() {
	MaxSortablePrecDec = OnePrecDec().Quo(SmallestPrecDec())
}

// ValidSortablePrecDec ensures that a PrecDec is within the sortable bounds,
// a PrecDec can't have a precision of less than 10^-18.
// Max sortable decimal was set to the reciprocal of SmallestPrecDec.
func ValidSortablePrecDec(dec PrecDec) bool {
	return dec.Abs().LTE(MaxSortablePrecDec)
}

// SortablePrecDecBytes returns a byte slice representation of a PrecDec that can be sorted.
// Left and right pads with 0s so there are 18 digits to left and right of the decimal point.
// For this reason, there is a maximum and minimum value for this, enforced by ValidSortablePrecDec.
func SortablePrecDecBytes(dec PrecDec) []byte {
	if !ValidSortablePrecDec(dec) {
		panic("dec must be within bounds")
	}
	// Instead of adding an extra byte to all sortable decs in order to handle max sortable, we just
	// makes its bytes be "max" which comes after all numbers in ASCIIbetical order
	if dec.Equal(MaxSortablePrecDec) {
		return []byte("max")
	}
	// For the same reason, we make the bytes of minimum sortable dec be --, which comes before all numbers.
	if dec.Equal(MaxSortablePrecDec.Neg()) {
		return []byte("--")
	}
	// We move the negative sign to the front of all the left padded 0s, to make negative numbers come before positive numbers
	if dec.IsNegative() {
		return append([]byte("-"), fmt.Appendf(nil, fmt.Sprintf("%%0%ds", Precision*2+1), dec.Abs().String())...)
	}
	return fmt.Appendf(nil, fmt.Sprintf("%%0%ds", Precision*2+1), dec.String())
}

// reuse nil values
var nilJSON []byte

func init() {
	empty := new(big.Int)
	bz, _ := empty.MarshalText()
	nilJSON, _ = json.Marshal(string(bz))
}

// MarshalJSON marshals the decimal
func (d PrecDec) MarshalJSON() ([]byte, error) {
	if d.i == nil {
		return nilJSON, nil
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON defines custom decoding scheme
func (d *PrecDec) UnmarshalJSON(bz []byte) error {
	if d.i == nil {
		d.i = new(big.Int)
	}

	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}

	// TODO: Reuse dec allocation
	newPrecDec, err := NewPrecDecFromStr(text)
	if err != nil {
		return err
	}

	d.i = newPrecDec.i
	return nil
}

// MarshalYAML returns the YAML representation.
func (d PrecDec) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// Marshal implements the gogo proto custom type interface.
func (d PrecDec) Marshal() ([]byte, error) {
	i := d.i
	if i == nil {
		i = new(big.Int)
	}
	return i.MarshalText()
}

// MarshalTo implements the gogo proto custom type interface.
func (d *PrecDec) MarshalTo(data []byte) (n int, err error) {
	i := d.i
	if i == nil {
		i = new(big.Int)
	}

	if i.Cmp(zeroInt) == 0 {
		copy(data, []byte{0x30})
		return 1, nil
	}

	bz, err := d.Marshal()
	if err != nil {
		return 0, err
	}

	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (d *PrecDec) Unmarshal(data []byte) error {
	if len(data) == 0 {
		d = nil
		return nil
	}

	if d.i == nil {
		d.i = new(big.Int)
	}

	if err := d.i.UnmarshalText(data); err != nil {
		return err
	}

	if d.i.BitLen() > maxPrecDecBitLen {
		return fmt.Errorf("decimal out of range; got: %d, max: %d", d.i.BitLen(), maxPrecDecBitLen)
	}

	return nil
}

// Size implements the gogo proto custom type interface.
func (d *PrecDec) Size() int {
	bz, _ := d.Marshal()
	return len(bz)
}

// Override Amino binary serialization by proxying to protobuf.
func (d PrecDec) MarshalAmino() ([]byte, error)   { return d.Marshal() }
func (d *PrecDec) UnmarshalAmino(bz []byte) error { return d.Unmarshal(bz) }

// helpers

// test if two decimal arrays are equal
func PrecDecsEqual(d1s, d2s []PrecDec) bool {
	if len(d1s) != len(d2s) {
		return false
	}

	for i, d1 := range d1s {
		if !d1.Equal(d2s[i]) {
			return false
		}
	}
	return true
}

// minimum decimal between two
func MinPrecDec(d1, d2 PrecDec) PrecDec {
	if d1.LT(d2) {
		return d1
	}
	return d2
}

// maximum decimal between two
func MaxPrecDec(d1, d2 PrecDec) PrecDec {
	if d1.LT(d2) {
		return d2
	}
	return d1
}

// intended to be used with require/assert:  require.True(PrecDecEq(...))
func PrecDecEq(t *testing.T, exp, got PrecDec) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func PrecDecApproxEq(t *testing.T, d1, d2, tol PrecDec) (*testing.T, bool, string, string, string) {
	diff := d1.Sub(d2).Abs()
	return t, diff.LTE(tol), "expected |d1 - d2| <:\t%v\ngot |d1 - d2| = \t\t%v", tol.String(), diff.String()
}

func hasOnlyDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// FormatPrecDec formats a decimal (as encoded in protobuf) into a value-rendered
// string following ADR-050. This function operates with string manipulation
// (instead of manipulating the sdk.PrecDec object).
func FormatPrecDec(v string) (string, error) {
	parts := strings.Split(v, ".")
	if len(parts) > 2 {
		return "", fmt.Errorf("invalid decimal: too many points in %s", v)
	}

	intPart, err := math.FormatInt(parts[0])
	if err != nil {
		return "", err
	}

	if len(parts) == 1 {
		return intPart, nil
	}

	decPart := strings.TrimRight(parts[1], "0")
	if len(decPart) == 0 {
		return intPart, nil
	}

	// Ensure that the decimal part has only digits.
	// https://github.com/cosmos/cosmos-sdk/issues/12811
	if !hasOnlyDigits(decPart) {
		return "", fmt.Errorf("non-digits detected after decimal point in: %q", decPart)
	}

	return intPart + "." + decPart, nil
}
