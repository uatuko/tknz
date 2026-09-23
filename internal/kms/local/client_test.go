package local

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientDecrypt(t *testing.T) {
	c, _ := newClientT(t)

	t.Run("round-trip", func(t *testing.T) {
		plaintext := []byte("s3cr3t")

		data, err := c.encrypt(plaintext)
		require.NoError(t, err)

		result, err := c.Decrypt(t.Context(), data)
		require.NoError(t, err)
		assert.Equal(t, plaintext, result)
	})

	t.Run("empty plaintext", func(t *testing.T) {
		data, err := c.encrypt(nil)
		require.NoError(t, err)

		result, err := c.Decrypt(t.Context(), data)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("tampered ciphertext", func(t *testing.T) {
		data, err := c.encrypt([]byte("s3cr3t"))
		require.NoError(t, err)

		data[len(data)-1] ^= 0xff

		_, err = c.Decrypt(t.Context(), data)
		assert.Error(t, err)
	})

	t.Run("tampered nonce", func(t *testing.T) {
		data, err := c.encrypt([]byte("s3cr3t"))
		require.NoError(t, err)

		data[0] ^= 0xff

		_, err = c.Decrypt(t.Context(), data)
		assert.Error(t, err)
	})

	t.Run("ciphertext from another key", func(t *testing.T) {
		data, err := c.encrypt([]byte("s3cr3t"))
		require.NoError(t, err)

		other, _ := newClientT(t)
		_, err = other.Decrypt(t.Context(), data)
		assert.Error(t, err)
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		_, err := c.Decrypt(t.Context(), make([]byte, c.aead.NonceSize()-1))
		assert.ErrorContains(t, err, "ciphertext too short")
	})

	t.Run("nonce only", func(t *testing.T) {
		// Long enough to pass the length check but with nothing to authenticate.
		_, err := c.Decrypt(t.Context(), make([]byte, c.aead.NonceSize()))
		assert.Error(t, err)
	})
}

func TestClientSign(t *testing.T) {
	c, key := newClientT(t)
	data := []byte("data to sign")

	t.Run("valid signature", func(t *testing.T) {
		sig, err := c.Sign(t.Context(), "unused-kms-key", "unused-kms-key-version", data)
		require.NoError(t, err)

		// Callers (internal/srv/authn.go, internal/srv/oidc/token.go) split the
		// signature at a fixed half way point to recover r and s.
		require.Len(t, sig, 2*((key.Curve.Params().BitSize+7)/8))

		r := new(big.Int).SetBytes(sig[:len(sig)/2])
		s := new(big.Int).SetBytes(sig[len(sig)/2:])

		digest := sha256.Sum256(data)
		assert.True(t, ecdsa.Verify(&key.PublicKey, digest[:], r, s))

		digest = sha256.Sum256([]byte("other data"))
		assert.False(t, ecdsa.Verify(&key.PublicKey, digest[:], r, s))
	})

	t.Run("signature is not reused", func(t *testing.T) {
		lhs, err := c.Sign(t.Context(), "unused-kms-key", "unused-kms-key-version", data)
		require.NoError(t, err)

		rhs, err := c.Sign(t.Context(), "unused-kms-key", "unused-kms-key-version", data)
		require.NoError(t, err)

		assert.NotEqual(t, lhs, rhs)
	})

	t.Run("cannot verify using a different key", func(t *testing.T) {
		other, _ := newClientT(t)

		sig, err := other.Sign(t.Context(), "unused-kms-key", "unused-kms-key-version", data)
		require.NoError(t, err)

		r := new(big.Int).SetBytes(sig[:len(sig)/2])
		s := new(big.Int).SetBytes(sig[len(sig)/2:])

		digest := sha256.Sum256(data)
		assert.False(t, ecdsa.Verify(&key.PublicKey, digest[:], r, s))
	})
}

// newClientT returns a client backed by a newly generated P-256 key, along
// with the key to verify signatures against.
func newClientT(t *testing.T) (*client, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	c, err := NewClient(t.Context(), writeEcKey(t, key))
	require.NoError(t, err)

	return c, key
}
