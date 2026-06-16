package jwk

import (
	"context"

	"github.com/felk-ai/idaas/internal/jose/jwa"
)

const (
	KtyEc  Kty = "EC"
	KtyRsa Kty = "RSA"

	UseSig Use = "sig"

	KeyOpSign   KeyOp = "sign"
	KeyOpVerify KeyOp = "verify"

	cacheExpMins     = 20
	cacheExpRandMins = 10
)

type Kty string
type Use string
type KeyOp string

type Jwk struct {
	Alg    jwa.Alg `json:"alg,omitempty"`
	Kid    string  `json:"kid,omitempty"`
	Kty    Kty     `json:"kty,omitempty"`
	Use    Use     `json:"use,omitempty"`
	KeyOps []KeyOp `json:"key_ops,omitempty"`

	// Elliptic curve (EC) params
	Crv jwa.EcCrv `json:"crv,omitempty"`
	X   string    `json:"x,omitempty"`
	Y   string    `json:"y,omitempty"`

	// RSA (RSA) params
	N string `json:"n,omitempty"` // modulus
	E string `json:"e,omitempty"` // exponent
}

func Fetch(ctx context.Context, jwksUri string) ([]Jwk, error) {
	keys, err := cacheFetch(ctx, jwksUri)
	if err != nil {
		return nil, err
	}

	return keys, nil
}
