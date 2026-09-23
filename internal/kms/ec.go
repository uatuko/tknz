package kms

import (
	"context"
	"crypto/elliptic"
	"fmt"
	"math/big"
)

type EcKey struct {
	kid           string
	kmsKey        string
	kmsKeyVersion string
}

func (k *EcKey) Kid() string {
	return k.kid
}

func (k *EcKey) Sign(ctx context.Context, data []byte) ([]byte, error) {
	return client.Sign(ctx, k.kmsKey, k.kmsKeyVersion, data)
}

type EcSig struct {
	R *big.Int
	S *big.Int
}

func NewEcKey(kid string, kmsKey string, kmsKeyVersion string) *EcKey {
	return &EcKey{
		kid:           kid,
		kmsKey:        kmsKey,
		kmsKeyVersion: kmsKeyVersion,
	}
}

// ParseEcSig parses a signature made up of the fixed-width r and s coordinates
// concatenated together, as produced by [EcKey.Sign] and used by JOSE (ES256
// and friends).
func ParseEcSig(curve elliptic.Curve, sig []byte) (*EcSig, error) {
	size := ecCoordinateSize(curve)
	if len(sig) != 2*size {
		// FIXME: errors
		return nil, fmt.Errorf("invalid ec signature size (want: %v, have: %v)", 2*size, len(sig))
	}

	return &EcSig{
		R: new(big.Int).SetBytes(sig[:size]),
		S: new(big.Int).SetBytes(sig[size:]),
	}, nil
}

// ecCoordinateSize returns the byte size of an ec signature coordinate (r or s)
// on curve.
func ecCoordinateSize(curve elliptic.Curve) int {
	return (curve.Params().BitSize + 7) / 8
}
