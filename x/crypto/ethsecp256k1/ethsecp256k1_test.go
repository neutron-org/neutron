package ethsecp256k1

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func TestPrivKey(t *testing.T) {
	// validate type and equality
	privKeyBz, err := hex.DecodeString("d9b18131efa344763bd5e3d1f7c9a12bdd3b62adf178fd25ec01b3594226b2d3")
	require.NoError(t, err)
	privKey := &PrivKey{
		Key: privKeyBz,
	}

	require.Implements(t, (*cryptotypes.PrivKey)(nil), privKey)

	// validate inequality
	privKey2 := GenerateKey()
	require.False(t, privKey.Equals(privKey2))

	// validate Ethereum address equality
	addr := privKey.PubKey().Address()
	require.NoError(t, err)

	expectedAddr, err := hex.DecodeString("ff4a64bddd522d3559b7dc2baa2de5364a7bc1d4")
	require.NoError(t, err)
	require.Equal(t, addr.Bytes(), expectedAddr)

	// validate we can sign some bytes
	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", 11, "hello world"))
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	require.Equal(t, hex.EncodeToString(sig), "351f94bfeacbce8c6aa1dc8f9aaa81e0f984df0352b41233b99c4576e486eb537471f3da6f62865e2f6720ea9a08e7aadb7d2d705f9879db0b5d5c0734f3b49f1b")
}

func TestPrivKey_PubKey(t *testing.T) {
	privKey := GenerateKey()

	// validate type and equality
	pubKey := &PubKey{
		Key: privKey.PubKey().Bytes(),
	}
	require.Implements(t, (*cryptotypes.PubKey)(nil), pubKey)

	// validate inequality
	privKey2 := GenerateKey()
	require.False(t, pubKey.Equals(privKey2.PubKey()))

	// validate signature
	msg := []byte("hello world")
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	res := pubKey.VerifySignature(msg, sig)
	require.True(t, res)
}

func TestVerifySignatureECDSA(t *testing.T) {
	// Given data (message, signature, and public key)
	msg := "Hello, MetaMask!"
	// msgHash := keccak256([]byte(msg))
	formattedMsg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	sigHex := "3dadd2820aad62a5e545f7b18178708f0f63afd667f9d2535e43870b52e57a1f333f416aa9485deaed22ac1e2ffe35485afdb45989c797acd654b73a881ace1f1b"
	pubKeyHex := "044c352d52ba4e507085205e9a029432defbc8d8f05ed828cbce0eb1a8823097723dc9caa6c60c17ad9073c2cdcdb409fe20110c40359607a64ca22d6607770655"
	// Decode signature and public key from hex
	sigBytes, err := hex.DecodeString(sigHex)
	require.NoError(t, err)
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	require.NoError(t, err)
	// Create pubKey struct
	pubKey := PubKey{
		Key: pubKeyBytes,
	}

	require.True(t, pubKey.VerifySignature([]byte(formattedMsg), sigBytes), "Valid signature should pass verification")
}

func TestPubKeyAddressUncompressed(t *testing.T) {
	// Given valid public key (uncompressed 65-byte key starting with 0x04)
	pubKeyHex := "0404794d0d9aa382bb479bf05ef71c1527af06f649f2fa659f83e08b602b8fba48e2ef4c82ed6d77487e56e9e89a55785f2ae3e4a84f4eee8295ff4cde1e5c55a9"
	expectedAddress := "A53AEF059604AD5B48DCA84E60F59CDACCF61C45" // Expected Ethereum address

	// Decode hex string into bytes
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	require.NoError(t, err)

	// Create PubKey struct, directly injecting the public key bytes
	pubKey := PubKey{Key: pubKeyBytes}

	// Compute address
	computedAddress := pubKey.Address().String() // fails here!
	require.Equal(t, expectedAddress, computedAddress, "Derived address should match expected Ethereum address")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid public key length, but function did not panic")
		}
	}()
	invalidPubKey := PubKey{Key: []byte("12345678901234567890")} // Incorrect length,
	invalidPubKey.Address()                                      // Should panic
}

func TestMarshalAmino(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	privKey := GenerateKey()

	pubKey := privKey.PubKey().(*PubKey)

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		typ       interface{}
		expBinary []byte
		expJSON   string
	}{
		{
			"ethsecp256k1 private key",
			privKey,
			&PrivKey{},
			append([]byte{32}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"ethsecp256k1 public key",
			pubKey,
			&PubKey{},
			append([]byte{33}, pubKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(pubKey.Bytes()) + "\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Do a round trip of encoding/decoding binary.
			bz, err := aminoCdc.Marshal(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expBinary, bz)

			err = aminoCdc.Unmarshal(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)

			// Do a round trip of encoding/decoding JSON.
			bz, err = aminoCdc.MarshalJSON(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expJSON, string(bz))

			err = aminoCdc.UnmarshalJSON(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)
		})
	}
}
