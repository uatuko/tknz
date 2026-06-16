package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"slices"
	"time"
)

const (
	joseHeaderAlgES256 = "ES256"
	joseHeaderTypJwt   = "JWT"

	pemBlockTypeEcPrivateKey = "EC PRIVATE KEY"
)

type joseHeader struct {
	Alg string `json:"alg,omitempty"`
	Kid string `json:"kid,omitempty"`
	Typ string `json:"typ,omitempty"`
}

type jwtClaims struct {
	Aud string `json:"aud,omitempty"`
	Exp int64  `json:"exp,omitempty"`
	Iat int64  `json:"iat,omitempty"`
	Iss string `json:"iss,omitempty"`
	Jti string `json:"jti,omitempty"`
	Nbf int64  `json:"nbf,omitempty"`
	Sub string `json:"sub,omitempty"`
}

func main() {
	key := flag.String("key", "", "pem encoded signing key (e.g. conf/key.pem)")
	kid := flag.String("kid", "", "signing key id")
	aud := flag.String("aud", "", "audience claim")
	sub := flag.String("sub", "", "subject claim")
	iss := flag.String("iss", "", "issuer claim")
	flag.Parse()

	pemKey, err := os.ReadFile(*key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] failed to read signing key, err: %v\n", err)
		os.Exit(1)
	}

	block, _ := pem.Decode(pemKey)
	if block == nil {
		fmt.Fprintf(os.Stderr, "[error] not pem encoded key found is file %v\n", *key)
		os.Exit(1)
	}

	if block.Type != pemBlockTypeEcPrivateKey {
		fmt.Fprintf(os.Stderr, "[error] unsupported pem encoded key (want: '%v', have '%v')\n", pemBlockTypeEcPrivateKey, block.Type)
		os.Exit(1)
	}

	header := joseHeader{
		Alg: joseHeaderAlgES256,
		Kid: *kid,
		Typ: joseHeaderTypJwt,
	}

	now := time.Now().UTC()
	claims := jwtClaims{
		Aud: *aud,
		Exp: now.Add(30 * time.Minute).Unix(),
		Iat: now.Unix(),
		Iss: *iss,
		Sub: *sub,
	}

	payload := fmt.Sprintf("%s.%s", jwsEncode(header), jwsEncode(claims))
	sig, err := jwsSign(payload, block)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] failed to create jws signature, err: %v\n", err)
		os.Exit(1)
	}

	token := fmt.Sprintf("%s.%s", payload, sig)
	fmt.Println(token)
}

func jwsEncode(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func jwsSign(payload string, block *pem.Block) (string, error) {
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256([]byte(payload))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		return "", nil
	}

	return base64.RawURLEncoding.EncodeToString(slices.Concat(r.Bytes(), s.Bytes())), nil
}
