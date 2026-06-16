package jws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/felk-ai/idaas/internal/jose/jwa"
	"github.com/felk-ai/idaas/internal/jose/jwk"
)

type JwkFunc func(ctx context.Context, kid string) (*jwk.Jwk, error)

type Jws struct {
	b64hdr     string
	b64payload string
	b64sig     string
	hdr        Header
}

func (jws *Jws) Header() Header {
	return jws.hdr
}

func (jws *Jws) Payload() string {
	return jws.b64payload
}

func (jws *Jws) Validate(ctx context.Context, jwkFn JwkFunc) error {
	k, err := jwkFn(ctx, jws.hdr.Kid)
	if err != nil {
		return fmt.Errorf("invalid jws (%w)", err)
	}

	if k.Alg != jws.hdr.Alg {
		return fmt.Errorf("invalid jws, key algorithm mismatch")
	}

	err = jws.verify(k)
	if err != nil {
		return fmt.Errorf("invalid jws (%w)", err)
	}

	return nil
}

func (jws *Jws) signingInput() string {
	return fmt.Sprintf("%s.%s", jws.b64hdr, jws.b64payload)
}

func (jws *Jws) verify(k *jwk.Jwk) error {
	if k.Use != jwk.UseSig {
		return fmt.Errorf("jwk cannot be used for verifying signatures")
	}

	switch k.Alg {
	case jwa.AlgRS256:
		return verifyRSA256(k, jws.signingInput(), jws.b64sig)
	default:
		return fmt.Errorf("unknown jwk algorithm")
	}
}

func Deserialize(compact string) (*Jws, error) {
	// JOSE header
	b64hdr, remain, ok := strings.Cut(compact, ".")
	if !ok {
		return nil, fmt.Errorf("invalid jws compact serialization")
	}

	var hdr Header
	err := decode(b64hdr, &hdr)
	if err != nil {
		return nil, fmt.Errorf("invalid jws compact serialization (%w)", err)
	}

	// JWS payload
	b64payload, remain, ok := strings.Cut(remain, ".")
	if !ok {
		return nil, fmt.Errorf("invalid jws compact serialization")
	}

	// JWS signature
	var b64sig string
	if hdr.Alg == jwa.AlgNone {
		if remain != "" {
			return nil, fmt.Errorf("invalid jws compact serialization")
		}
	} else {
		var unexpected bool
		b64sig, _, unexpected = strings.Cut(remain, ".")
		if unexpected {
			return nil, fmt.Errorf("invalid jws compact serialization")
		}
	}

	return &Jws{
		b64hdr:     b64hdr,
		b64payload: b64payload,
		b64sig:     b64sig,
		hdr:        hdr,
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
