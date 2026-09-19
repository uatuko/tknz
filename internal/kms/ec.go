package kms

import (
	"context"
	"math/big"
)

type EcSig struct {
	R *big.Int
	S *big.Int
}

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

func NewEcKey(kid string, kmsKey string, kmsKeyVersion string) *EcKey {
	return &EcKey{
		kid:           kid,
		kmsKey:        kmsKey,
		kmsKeyVersion: kmsKeyVersion,
	}
}
