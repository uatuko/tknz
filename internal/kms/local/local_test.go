package local

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("P-256 key", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		c, err := NewClient(t.Context(), writeEcKey(t, key))
		require.NoError(t, err)

		assert.Equal(t, key, c.ec)
		assert.NotNil(t, c.aead)
	})

	t.Run("P-384 key", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		require.NoError(t, err)

		c, err := NewClient(t.Context(), writeEcKey(t, key))
		require.NoError(t, err)

		assert.Equal(t, key, c.ec)
		assert.NotNil(t, c.aead)
	})

	t.Run("no key file", func(t *testing.T) {
		_, err := NewClient(t.Context(), filepath.Join(t.TempDir(), "nonexistent.pem"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("not pem encoded", func(t *testing.T) {
		fname := filepath.Join(t.TempDir(), "key.pem")
		require.NoError(t, os.WriteFile(fname, []byte("not a pem file"), 0o600))

		_, err := NewClient(t.Context(), fname)
		assert.ErrorContains(t, err, "no pem encoded key found")
	})

	t.Run("unsupported pem block type", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		b, err := x509.MarshalPKCS8PrivateKey(key)
		require.NoError(t, err)

		fname := filepath.Join(t.TempDir(), "key.pem")
		require.NoError(t, os.WriteFile(
			fname, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b}), 0o600))

		_, err = NewClient(t.Context(), fname)
		assert.ErrorContains(t, err, "unsupported pem encoded key")
	})

	t.Run("malformed key", func(t *testing.T) {
		fname := filepath.Join(t.TempDir(), "key.pem")
		require.NoError(t, os.WriteFile(fname, pem.EncodeToMemory(&pem.Block{
			Type:  pemBlockTypeEcPrivateKey,
			Bytes: []byte("not a der encoded key"),
		}), 0o600))

		_, err := NewClient(t.Context(), fname)
		assert.Error(t, err)
	})

	t.Run("symmetric key is stable across clients", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		fname := writeEcKey(t, key)

		lhs, err := NewClient(t.Context(), fname)
		require.NoError(t, err)

		rhs, err := NewClient(t.Context(), fname)
		require.NoError(t, err)

		// Both clients derive the same symmetric key, i.e. data encrypted by one
		// can be decrypted from the other.
		data, err := lhs.encrypt([]byte("s3cr3t"))
		require.NoError(t, err)

		plaintext, err := rhs.Decrypt(t.Context(), data)
		require.NoError(t, err)
		assert.Equal(t, []byte("s3cr3t"), plaintext)
	})
}

// writeEcKey writes a pem encoded key to a file in a temporary directory and
// returns the file name.
func writeEcKey(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()

	b, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	fname := filepath.Join(t.TempDir(), "key.pem")
	require.NoError(t, os.WriteFile(
		fname, pem.EncodeToMemory(&pem.Block{Type: pemBlockTypeEcPrivateKey, Bytes: b}), 0o600))

	return fname
}
