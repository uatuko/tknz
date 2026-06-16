package jws

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"

	"go.tknz.dev/internal/jose/jwa"
	"go.tknz.dev/internal/jose/jwk"
)

func rsaPublicKey(k *jwk.Jwk) (*rsa.PublicKey, error) {
	if k.Kty != jwk.KtyRsa || k.Alg != jwa.AlgRS256 {
		return nil, fmt.Errorf("jwk is not rsa")
	}

	var n big.Int
	b, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("invalid jwk (%w)", err)
	}
	n.SetBytes(b)

	b, err = base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("invalid jwk (%w)", err)
	}

	if len(b) > 8 {
		return nil, fmt.Errorf("invalid jwk (rsa exponent too big)")
	}

	if len(b) < 8 {
		b = append(make([]byte, 8-len(b)), b...)
	}

	return &rsa.PublicKey{
		N: &n,
		E: int(binary.BigEndian.Uint64(b)),
	}, nil
}

func verifyRSA256(k *jwk.Jwk, input string, b64sig string) error {
	sig, err := base64.RawURLEncoding.DecodeString(b64sig)
	if err != nil {
		return fmt.Errorf("invalid signature (%w)", err)
	}

	pub, err := rsaPublicKey(k)
	if err != nil {
		return err
	}

	h := sha256.Sum256([]byte(input))
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig)
	if err != nil {
		return fmt.Errorf("signature verification failed (%w)", err)
	}

	return nil
}
