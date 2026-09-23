package local

import (
	"context"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"slices"
)

type client struct {
	aead cipher.AEAD
	ec   *ecdsa.PrivateKey
}

func (c *client) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	if len(data) < c.aead.NonceSize() {
		// FIXME: errors
		return nil, fmt.Errorf("decrypt: ciphertext too short")
	}

	nonce, ciphertext := data[:c.aead.NonceSize()], data[c.aead.NonceSize():]
	return c.aead.Open(nil, nonce, ciphertext, nil)
}

func (c *client) Sign(ctx context.Context, key string, keyVersion string, data []byte) ([]byte, error) {
	digest := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, c.ec, digest[:])
	if err != nil {
		return nil, err
	}

	size := (c.ec.Curve.Params().BitSize + 7) / 8
	return slices.Concat(
		r.FillBytes(make([]byte, size)),
		s.FillBytes(make([]byte, size)),
	), nil
}

func (c *client) encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return c.aead.Seal(nonce, nonce, data, nil), nil
}
