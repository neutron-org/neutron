package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

type ChecksumConst int

const (
	// Version0Const is the original constant used in the checksum
	// verification for bech32.
	Version0Const ChecksumConst = 1

	// VersionMConst is the new constant used for bech32m checksum
	// verification.
	VersionMConst ChecksumConst = 0x2bc830a3
)

// Version defines the current set of bech32 versions.
type Version uint8

const (
	// Version0 defines the original bech version.
	Version0 Version = iota

	// VersionM is the new bech32 version defined in BIP-350, also known as
	// bech32m.
	VersionM

	// VersionUnknown denotes an unknown bech version.
	VersionUnknown
)

// VersionToConsts maps bech32 versions to the checksum constant to be used
// when encoding, and asserting a particular version when decoding.
var VersionToConsts = map[Version]ChecksumConst{
	Version0: Version0Const,
	VersionM: VersionMConst,
}

// ConstsToVersion maps a bech32 constant to the version it's associated with.
var ConstsToVersion = map[ChecksumConst]Version{
	Version0Const: Version0,
	VersionMConst: VersionM,
}

// charset is the set of characters used in the data section of bech32 strings.
// Note that this is ordered, such that for a given charset[i], i is the binary
// value of the character.
const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// gen encodes the generator polynomial for the bech32 BCH checksum.
var gen = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

func bech32Polymod(hrp string, values, checksum []byte) int {
	chk := 1

	// Account for the high bits of the HRP in the checksum.
	for i := 0; i < len(hrp); i++ {
		b := chk >> 25
		hiBits := int(hrp[i]) >> 5
		chk = (chk&0x1ffffff)<<5 ^ hiBits
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}

	// Account for the separator (0) between high and low bits of the HRP.
	// x^0 == x, so we eliminate the redundant xor used in the other rounds.
	b := chk >> 25
	chk = (chk & 0x1ffffff) << 5
	for i := 0; i < 5; i++ {
		if (b>>uint(i))&1 == 1 {
			chk ^= gen[i]
		}
	}

	// Account for the low bits of the HRP.
	for i := 0; i < len(hrp); i++ {
		b := chk >> 25
		loBits := int(hrp[i]) & 31
		chk = (chk&0x1ffffff)<<5 ^ loBits
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}

	// Account for the values.
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ int(v)
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}

	if checksum == nil {
		// A nil checksum is used during encoding, so assume all bytes are zero.
		// x^0 == x, so we eliminate the redundant xor used in the other rounds.
		for v := 0; v < 6; v++ {
			b := chk >> 25
			chk = (chk & 0x1ffffff) << 5
			for i := 0; i < 5; i++ {
				if (b>>uint(i))&1 == 1 {
					chk ^= gen[i]
				}
			}
		}
	} else {
		// Checksum is provided during decoding, so use it.
		for _, v := range checksum {
			b := chk >> 25
			chk = (chk&0x1ffffff)<<5 ^ int(v)
			for i := 0; i < 5; i++ {
				if (b>>uint(i))&1 == 1 {
					chk ^= gen[i]
				}
			}
		}
	}

	return chk
}

func writeBech32Checksum(hrp string, data []byte, bldr *strings.Builder,
	version Version) {

	bech32Const := int(VersionToConsts[version])
	polymod := bech32Polymod(hrp, data, nil) ^ bech32Const
	for i := 0; i < 6; i++ {
		b := byte((polymod >> uint(5*(5-i))) & 31)

		// This can't fail, given we explicitly cap the previous b byte by the
		// first 31 bits.
		c := charset[b]
		bldr.WriteByte(c)
	}
}

func encodeGeneric(hrp string, data []byte,
	version Version) (string, error) {

	// The resulting bech32 string is the concatenation of the lowercase
	// hrp, the separator 1, data and the 6-byte checksum.
	hrp = strings.ToLower(hrp)
	var bldr strings.Builder
	bldr.Grow(len(hrp) + 1 + len(data) + 6)
	bldr.WriteString(hrp)
	bldr.WriteString("1")

	// Write the data part, using the bech32 charset.
	for _, b := range data {
		if int(b) >= len(charset) {
			return "someshit", nil
		}
		bldr.WriteByte(charset[b])
	}

	// Calculate and write the checksum of the data.
	writeBech32Checksum(hrp, data, &bldr, version)

	return bldr.String(), nil
}

func Encode(hrp string, data []byte) (string, error) {
	return encodeGeneric(hrp, data, Version0)
}

func ConvertBits(data []byte, fromBits, toBits uint8, pad bool) ([]byte, error) {
	if fromBits < 1 || fromBits > 8 || toBits < 1 || toBits > 8 {
		return nil, nil
	}

	// Determine the maximum size the resulting array can have after base
	// conversion, so that we can size it a single time. This might be off
	// by a byte depending on whether padding is used or not and if the input
	// data is a multiple of both fromBits and toBits, but we ignore that and
	// just size it to the maximum possible.
	maxSize := len(data)*int(fromBits)/int(toBits) + 1

	// The final bytes, each byte encoding toBits bits.
	regrouped := make([]byte, 0, maxSize)

	// Keep track of the next byte we create and how many bits we have
	// added to it out of the toBits goal.
	nextByte := byte(0)
	filledBits := uint8(0)

	for _, b := range data {

		// Discard unused bits.
		b <<= 8 - fromBits

		// How many bits remaining to extract from the input data.
		remFromBits := fromBits
		for remFromBits > 0 {
			// How many bits remaining to be added to the next byte.
			remToBits := toBits - filledBits

			// The number of bytes to next extract is the minimum of
			// remFromBits and remToBits.
			toExtract := remFromBits
			if remToBits < toExtract {
				toExtract = remToBits
			}

			// Add the next bits to nextByte, shifting the already
			// added bits to the left.
			nextByte = (nextByte << toExtract) | (b >> (8 - toExtract))

			// Discard the bits we just extracted and get ready for
			// next iteration.
			b <<= toExtract
			remFromBits -= toExtract
			filledBits += toExtract

			// If the nextByte is completely filled, we add it to
			// our regrouped bytes and start on the next byte.
			if filledBits == toBits {
				regrouped = append(regrouped, nextByte)
				filledBits = 0
				nextByte = 0
			}
		}
	}

	// We pad any unfinished group if specified.
	if pad && filledBits > 0 {
		nextByte <<= toBits - filledBits
		regrouped = append(regrouped, nextByte)
		filledBits = 0
		nextByte = 0
	}

	// Any incomplete group must be <= 4 bits, and all zeroes.
	if filledBits > 0 && (filledBits > 4 || nextByte != 0) {
		return nil, nil
	}

	return regrouped, nil
}

func EncodeFromBase256(hrp string, data []byte) (string, error) {
	converted, err := ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", err
	}
	return Encode(hrp, converted)
}

func main() {
	codeId, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		panic(err)
	}
	instId, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		panic(err)
	}

	contractID := make([]byte, 16)
	binary.BigEndian.PutUint64(contractID[:8], codeId)
	binary.BigEndian.PutUint64(contractID[8:], instId)

	mKey := append([]byte("wasm"), 0)

	s := "module"
	hasher := sha256.New()
	var buf []byte
	sHdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	bufHdr.Data = sHdr.Data
	bufHdr.Cap = sHdr.Len
	bufHdr.Len = sHdr.Len
	_, err = hasher.Write(buf)
	if err != nil {
		panic(err)
	}

	key := append(mKey, contractID...)
	th := hasher.Sum(nil)
	hasher.Reset()
	_, err = hasher.Write(th)
	if err != nil {
		panic(err)
	}
	_, err = hasher.Write(key)
	if err != nil {
		panic(err)
	}
	result := hasher.Sum(nil)[:32]

	encoded, err := EncodeFromBase256("neutron", result)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Show the encoded data.
	fmt.Println(encoded)
}
