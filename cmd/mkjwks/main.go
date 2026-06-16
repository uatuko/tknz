package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	jwkAlgES256  = "ES256"
	jwkKtyEc     = "EC"
	jwkEcCrvP256 = "P-256"

	pemBlockTypeEcPrivateKey = "EC PRIVATE KEY"
)

type jwk struct {
	Alg    string   `json:"alg,omitempty"`
	Kid    string   `json:"kid,omitempty"`
	Kty    string   `json:"kty,omitempty"`
	Use    string   `json:"use,omitempty"`
	KeyOps []string `json:"key_ops,omitempty"`

	// Elliptic curve (EC) params
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`

	// RSA (RSA) params
	N string `json:"n,omitempty"` // modulus
	E string `json:"e,omitempty"` // exponent
}

func main() {
	keys := flag.String("keys", "", "comma separated list of pem encoded keys (e.g. conf/key.pem)")
	flag.Parse()

	ecKeys := make([]*ecdsa.PrivateKey, 0)
	for fname := range strings.SplitSeq(*keys, ",") {
		fname = strings.TrimSpace(fname)

		b, err := os.ReadFile(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] failed to read key from file, err: %v\n", err)
			os.Exit(1)
		}

		block, _ := pem.Decode(b)
		if block.Type != pemBlockTypeEcPrivateKey {
			fmt.Fprintf(os.Stderr, "[error] unsupported pem encoded key (want: '%v', have '%v')\n", pemBlockTypeEcPrivateKey, block.Type)
			os.Exit(1)
		}

		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] failed to parse key, err: %v\n", err)
		}

		ecKeys = append(ecKeys, key)
	}

	jwks := make([]jwk, 0, len(ecKeys))
loop:
	for _, ec := range ecKeys {
		k := jwk{
			Kty: jwkKtyEc,
			X:   base64.RawURLEncoding.EncodeToString(ec.X.Bytes()),
			Y:   base64.RawURLEncoding.EncodeToString(ec.Y.Bytes()),
		}

		switch ec.Curve {
		case elliptic.P256():
			k.Alg = jwkAlgES256
			k.Crv = jwkEcCrvP256
		default:
			fmt.Fprintf(os.Stderr, "[warn] unsupported ec curve, ignoring key\n")
			break loop
		}

		jwks = append(jwks, k)
	}

	b, err := json.Marshal(jwks)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", b)
}
