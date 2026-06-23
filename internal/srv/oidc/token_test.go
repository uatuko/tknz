package oidc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.tknz.dev/internal/db"
)

func TestJwsValidate(t *testing.T) {
	// Ref: https://www.rfc-editor.org/rfc/rfc7515.html#appendix-A.3
	t.Run("ES256", func(t *testing.T) {
		jwk := db.JwkParams{
			Kty: db.JwkKtyEc,
			Crv: db.JwkEcCrvP256,
			X:   "f83OJ3D2xF1Bg8vub9tLe1gHMzV76e8Tus9uPHvRVEU",
			Y:   "x_FEzRu9m36HLN_tue659LNpXW6pCyStikYjKIWI5a0",
		}

		key, err := jwk.EcPublicKey()
		require.NoError(t, err)

		payload := strings.ReplaceAll(`
eyJhbGciOiJFUzI1NiJ9
.
eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFt
cGxlLmNvbS9pc19yb290Ijp0cnVlfQ
`, "\n", "")

		sig := strings.ReplaceAll(`
DtEhU3ljbEg8L38VWAfUAqOyKAM6-Xx-F4GawxaepmXFCgfTjDxw5djxLa8ISlSA
pmWQxfKTUJqPP3-Kg6NU1Q
`, "\n", "")

		require.NoError(t, jwsValidate(payload, sig, key))
	})
}

func TestJwtValidate(t *testing.T) {
	const header = "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9" // {"alg":"ES256","typ":"JWT"}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	now := time.Now().UTC()
	claims := jwtClaims{
		Exp: now.Add(1 * time.Minute).Unix(),
		Iat: now.Unix(),
	}
	b, err := json.Marshal(claims)
	require.NoError(t, err)
	payload := fmt.Sprintf("%s.%s", header, base64.RawURLEncoding.EncodeToString(b))

	keys := []db.JwkParams{
		{
			Kty: db.JwkKtyEc,
			Crv: db.JwkEcCrvP256,
			X:   base64.RawURLEncoding.EncodeToString(key.X.Bytes()),
			Y:   base64.RawURLEncoding.EncodeToString(key.Y.Bytes()),
		},
	}

	digest := sha256.Sum256([]byte(payload))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	require.NoError(t, err)

	token := fmt.Sprintf("%s.%s",
		payload, base64.RawURLEncoding.EncodeToString(slices.Concat(r.Bytes(), s.Bytes())))

	require.NoError(t, jwtValidate(token, jwtClaims{}, keys))
}
