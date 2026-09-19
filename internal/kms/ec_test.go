package kms

import (
	"crypto/elliptic"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEcCoordinateSize(t *testing.T) {
	assert.Equal(t, 32, ecCoordinateSize(elliptic.P256()))
	assert.Equal(t, 48, ecCoordinateSize(elliptic.P384()))
	assert.Equal(t, 66, ecCoordinateSize(elliptic.P521()))
}

func TestParseEcSig(t *testing.T) {
	t.Run("success: P-256", func(t *testing.T) {
		sig := make([]byte, 64)
		sig[31], sig[63] = 0x01, 0x02

		result, err := ParseEcSig(elliptic.P256(), sig)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(1), result.R)
		assert.Equal(t, big.NewInt(2), result.S)
	})

	t.Run("success: P-384", func(t *testing.T) {
		sig := make([]byte, 96)
		sig[47], sig[95] = 0x01, 0x02

		result, err := ParseEcSig(elliptic.P384(), sig)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(1), result.R)
		assert.Equal(t, big.NewInt(2), result.S)
	})

	t.Run("success: coordinates with leading zeros", func(t *testing.T) {
		// r and s are padded to the curve size, a coordinate small enough to have
		// leading zero bytes must still land on the right side of the split.
		r, s := big.NewInt(0xbeef), new(big.Int).SetBytes([]byte{0xf0, 0x0d})

		sig := make([]byte, 64)
		r.FillBytes(sig[:32])
		s.FillBytes(sig[32:])

		result, err := ParseEcSig(elliptic.P256(), sig)
		require.NoError(t, err)
		assert.Equal(t, r, result.R)
		assert.Equal(t, s, result.S)
	})

	t.Run("success: zero signature", func(t *testing.T) {
		result, err := ParseEcSig(elliptic.P256(), make([]byte, 64))
		require.NoError(t, err)
		assert.Zero(t, result.R.Sign())
		assert.Zero(t, result.S.Sign())
	})

	tests := []struct {
		name string
		sig  []byte
	}{
		{"error: nil", nil},
		{"error: empty", []byte{}},
		{"error: truncated coordinate", make([]byte, 63)},
		{"error: half a signature", make([]byte, 32)},
		{"error: too long", make([]byte, 65)},
		{"error: wrong curve size", make([]byte, 96)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseEcSig(elliptic.P256(), test.sig)
			assert.ErrorContains(t, err, "invalid ec signature size")
		})
	}
}
