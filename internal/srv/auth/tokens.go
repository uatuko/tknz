package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/jose/jwa"
	"go.tknz.dev/internal/jose/jws"
	"go.tknz.dev/internal/kms"
	"go.tknz.dev/internal/pb"
	"go.tknz.dev/internal/srv/common"
	"google.golang.org/protobuf/proto"
)

type idTokenClaims struct {
	Aud   string `json:"aud,omitempty"`
	Exp   int64  `json:"exp,omitempty"`
	Iat   int64  `json:"iat,omitempty"`
	Iss   string `json:"iss,omitempty"`
	Nonce string `json:"nonce,omitempty"`
	Sub   string `json:"sub,omitempty"`
}

func NewAccessToken(ctx context.Context, app *db.App, code *db.AuthCode) (string, error) {
	token, err := db.NewAuthToken(ctx, accessTokenTtl, db.AuthTokenAttrs{
		AppId:      app.Id(),
		ProviderId: code.Attrs().ProviderId,
		Aud:        app.Attrs().Aud,
		Sub:        code.Attrs().Sub,
	})
	if err != nil {
		return "", err
	}

	key, err := findEcKey(ctx)
	if err != nil {
		return "", err
	}

	tokenBytes := token.Token()
	ecSig, err := key.Sign(ctx, tokenBytes)
	if err != nil {
		return "", err
	}

	pbToken := pb.Token{
		Kid:       key.Kid(),
		Token:     tokenBytes,
		Signature: slices.Concat(ecSig.R.Bytes(), ecSig.S.Bytes()),
	}

	b, err := proto.Marshal(&pbToken)
	if err != nil {
		return "", err
	}

	return AccessTokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

func NewIdToken(ctx context.Context, code *db.AuthCode) (string, error) {
	header := jws.Header{
		Alg: jwa.AlgNone,
		Typ: jws.HeaderTypJWT,
	}

	now := time.Now().UTC()
	payload := jwsEncode(idTokenClaims{
		Aud: code.Attrs().ClientId,
		Exp: now.Add(idTokenTtl).Unix(),
		Iat: now.Unix(),
		Sub: code.Attrs().Sub,
	})

	return jwsSign(ctx, &header, payload)
}

func findEcKey(ctx context.Context) (*kms.EcKey, error) {
	jwks, err := db.ListJwks(ctx, common.SysSpaceId)
	if err != nil {
		return nil, err
	}

	for _, jwk := range jwks {
		if jwk.Params().Kty == db.JwkKtyEc && slices.Contains(jwk.Params().KeyOps, db.JwkKeyOpSign) {
			return kms.NewEcKey(jwk.Kid(), jwk.Attrs().KmsKey, jwk.Attrs().KmsKeyVersion), nil
		}
	}

	// FIXME: errors
	return nil, fmt.Errorf("not found")
}

func jwsEncode(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func jwsSign(ctx context.Context, header *jws.Header, payload string) (string, error) {
	key, err := findEcKey(ctx)
	if err != nil {
		return "", err
	}

	header.Alg = joseHeaderAlgES256
	header.Kid = key.Kid()
	input := fmt.Sprintf("%s.%s", jwsEncode(header), payload)

	ecSig, err := key.Sign(ctx, []byte(input))
	if err != nil {
		return "", err
	}

	jwsSig := base64.RawURLEncoding.EncodeToString(slices.Concat(ecSig.R.Bytes(), ecSig.S.Bytes()))

	return fmt.Sprintf("%s.%s", input, jwsSig), nil
}
