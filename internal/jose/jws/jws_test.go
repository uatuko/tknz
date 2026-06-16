package jws

import (
	"context"
	"strings"
	"testing"

	"github.com/felk-ai/idaas/internal/jose/jwa"
	"github.com/felk-ai/idaas/internal/jose/jwk"
	"github.com/stretchr/testify/require"
)

func TestDeserialize(t *testing.T) {
	// Ref: https://www.rfc-editor.org/rfc/rfc7515.html#appendix-A.5
	t.Run("unsecured", func(t *testing.T) {
		compact := strings.ReplaceAll(`
eyJhbGciOiJub25lIn0
.
eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFt
cGxlLmNvbS9pc19yb290Ijp0cnVlfQ
.
`, "\n", "")

		jws, err := Deserialize(compact)
		require.NoError(t, err)

		hdr := jws.Header()
		require.Equal(t, jwa.AlgNone, hdr.Alg)
		require.Empty(t, hdr.Kid)
		require.Empty(t, hdr.Typ)

		require.Equal(t, "eyJhbGciOiJub25lIn0", jws.b64hdr)
		require.Equal(t, "eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ", jws.Payload())
		require.Empty(t, jws.b64sig)
	})

	// Ref: https://www.rfc-editor.org/rfc/rfc7515.html#appendix-A.2
	t.Run("RS256", func(t *testing.T) {
		compact := strings.ReplaceAll(`
eyJhbGciOiJSUzI1NiJ9
.
eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFt
cGxlLmNvbS9pc19yb290Ijp0cnVlfQ
.
cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7
AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4
BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K
0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqv
hJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrB
p0igcN_IoypGlUPQGe77Rw`, "\n", "")

		jws, err := Deserialize(compact)
		require.NoError(t, err)

		hdr := jws.Header()
		require.Equal(t, jwa.AlgRS256, hdr.Alg)
		require.Empty(t, hdr.Kid)
		require.Empty(t, hdr.Typ)

		require.Equal(t, "eyJhbGciOiJSUzI1NiJ9", jws.b64hdr)
		require.Equal(t, "eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ", jws.Payload())

		sig := strings.ReplaceAll(`
cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7
AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4
BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K
0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqv
hJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrB
p0igcN_IoypGlUPQGe77Rw`, "\n", "")
		require.Equal(t, sig, jws.b64sig)
	})
}

func TestJwsValidate(t *testing.T) {
	// Ref: https://www.rfc-editor.org/rfc/rfc7515.html#appendix-A.2
	t.Run("RS256", func(t *testing.T) {
		jwkFn := func(ctx context.Context, kid string) (*jwk.Jwk, error) {
			return &jwk.Jwk{
				Alg: jwa.AlgRS256,
				Kty: jwk.KtyRsa,
				Use: jwk.UseSig,
				N: strings.ReplaceAll(`
ofgWCuLjybRlzo0tZWJjNiuSfb4p4fAkd_wWJcyQoTbji9k0l8W26mPddx
HmfHQp-Vaw-4qPCJrcS2mJPMEzP1Pt0Bm4d4QlL-yRT-SFd2lZS-pCgNMs
D1W_YpRPEwOWvG6b32690r2jZ47soMZo9wGzjb_7OMg0LOL-bSf63kpaSH
SXndS5z5rexMdbBYUsLA9e-KXBdQOS-UTo7WTBEMa2R2CapHg665xsmtdV
MTBQY4uDZlxvb3qCo5ZwKh9kG4LT6_I5IhlJH7aGhyxXFvUK-DWNmoudF8
NAco9_h9iaGNj8q2ethFkMLs91kzk2PAcDTW9gb54h4FRWyuXpoQ`, "\n", ""),
				E: "AQAB",
			}, nil
		}

		compact := strings.ReplaceAll(`
eyJhbGciOiJSUzI1NiJ9
.
eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFt
cGxlLmNvbS9pc19yb290Ijp0cnVlfQ
.
cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7
AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4
BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K
0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqv
hJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrB
p0igcN_IoypGlUPQGe77Rw`, "\n", "")

		jws, err := Deserialize(compact)
		require.NoError(t, err)

		err = jws.Validate(t.Context(), jwkFn)
		require.NoError(t, err)
	})
}
