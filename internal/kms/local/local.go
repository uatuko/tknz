package local

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const pemBlockTypeEcPrivateKey = "EC PRIVATE KEY"

func NewClient(ctx context.Context, fname string) (*client, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		// FIXME: errors
		return nil, fmt.Errorf("no pem encoded key found in file %v", fname)
	}

	if block.Type != pemBlockTypeEcPrivateKey {
		// FIXME: errors
		return nil, fmt.Errorf("unsupported pem encoded key (want: '%v', have: '%v')",
			pemBlockTypeEcPrivateKey, block.Type)
	}

	ec, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Derive a symmetric key from the signing key so a single key can be used for
	// both signing and encryption.
	dk := sha256.Sum256(block.Bytes)
	c, err := aes.NewCipher(dk[:])
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	return &client{
		aead: aead,
		ec:   ec,
	}, nil
}
