package jwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"go.tknz.dev/internal/jose/jwk"
	"go.tknz.dev/internal/jose/jws"
)

type JwkFunc func(ctx context.Context, kid string, iss string) (*jwk.Jwk, error)

type Jwt struct {
	claims Claims
	hdr    jws.Header
}

func (t *Jwt) Claims() *Claims {
	return &t.claims
}

func Validate(ctx context.Context, s string, jwkFn JwkFunc) (*Jwt, error) {
	sig, err := jws.Deserialize(s)
	if err != nil {
		return nil, fmt.Errorf("invalid jwt (%w)", err)
	}

	hdr := sig.Header()
	if hdr.Typ != jws.HeaderTypJWT {
		return nil, fmt.Errorf("invalid jwt")
	}

	// JWT claims
	var claims registeredClaims
	err = decode(sig.Payload(), &claims)
	if err != nil {
		return nil, fmt.Errorf("invalid jwt (%w)", err)
	}

	// TODO: private claims

	err = sig.Validate(ctx, mkJwkFunc(claims.Iss, jwkFn))
	if err != nil {
		return nil, fmt.Errorf("invalid jwt (%w)", err)
	}

	err = claims.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid jwt (%w)", err)
	}

	return &Jwt{
		claims: Claims{public: claims},
		hdr:    hdr,
	}, nil
}

func decode(b64 string, v any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(decoded), v)
	if err != nil {
		return err
	}

	return nil
}

func mkJwkFunc(iss string, jwkFn JwkFunc) jws.JwkFunc {
	return func(ctx context.Context, kid string) (*jwk.Jwk, error) {
		return jwkFn(ctx, kid, iss)
	}
}
