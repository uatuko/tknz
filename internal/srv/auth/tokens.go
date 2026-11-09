package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/jose/jwa"
	"go.tknz.dev/internal/jose/jws"
	"go.tknz.dev/internal/kms"
	"go.tknz.dev/internal/pb"
	"go.tknz.dev/internal/srv/common"
)

type idTokenClaims struct {
	Aud   string `json:"aud,omitempty"`
	Exp   int64  `json:"exp,omitempty"`
	Iat   int64  `json:"iat,omitempty"`
	Iss   string `json:"iss,omitempty"`
	Nonce string `json:"nonce,omitempty"`
	Sub   string `json:"sub,omitempty"`
}

// CheckAccessToken validates an access token issued by NewAccessToken and returns the identity and claims
// associated with it.
//
// Returns ErrInvalidAccessTokenPrefix if the token isn't an access token, ErrInvalidAccessToken if the
// token is invalid, or any other error for internal failures.
func CheckAccessToken(ctx context.Context, token string) (*db.Idn, *pb.Claims, error) {
	if !strings.HasPrefix(token, AccessTokenPrefix) {
		return nil, nil, ErrInvalidTokenPrefix
	}

	b, err := base64.RawURLEncoding.DecodeString(token[len(AccessTokenPrefix):])
	if err != nil {
		fmt.Printf("[error] failed to decode, err: %v\n", err)
		return nil, nil, ErrInvalidToken
	}

	var pbToken pb.Token
	err = proto.Unmarshal(b, &pbToken)
	if err != nil {
		fmt.Printf("[error] failed to unmarshal, err: %v\n", err)
		return nil, nil, ErrInvalidToken
	}

	jwk, err := db.RetrieveJwk(ctx, common.SysSpaceId, pbToken.GetKid())
	if err != nil {
		fmt.Printf("[error] failed to retrieve jwk, err: %v\n", err)
		return nil, nil, ErrInvalidToken
	}

	params := jwk.Params()
	key, err := params.EcPublicKey()
	if err != nil {
		fmt.Printf("[error] failed to retrieve ec public key, err: %v\n", err)
		return nil, nil, ErrServerError
	}

	sig, err := kms.ParseEcSig(key.Curve, pbToken.GetSignature())
	if err != nil {
		fmt.Printf("[error] invalid signature, err: %v\n", err)
		return nil, nil, ErrInvalidToken
	}

	b, err = proto.Marshal(pbToken.GetClaims())
	if err != nil {
		fmt.Printf("[error] failed to marshal claims, err: %v\n", err)
		return nil, nil, ErrServerError
	}

	b = append(b, pbToken.GetId()...)
	digest := sha256.Sum256(b)
	if !ecdsa.Verify(key, digest[:], sig.R, sig.S) {
		fmt.Println("[error] invalid signature")
		return nil, nil, ErrInvalidToken
	}

	idn, err := db.RetrieveIdnByAuthToken(ctx, pbToken.GetId())
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			fmt.Printf("[error] failed to retrieve idn, err: %v\n", err)
		}

		return nil, nil, ErrInvalidToken
	}

	claims := pbToken.GetClaims()
	if idn.Id() != claims.GetSub() {
		fmt.Printf("[error] token subject mismatch (want: %v, have: %v)\n", claims.GetSub(), idn.Id())
		return nil, nil, ErrInvalidToken
	}

	return idn, claims, nil
}

func NewAccessToken(ctx context.Context, app *db.App, code *db.AuthCode) (string, error) {
	attrs := db.AuthTokenAttrs{
		AppId:      app.Id(),
		ProviderId: code.ProviderId(),
		Aud:        app.Aud(),
		Sub:        code.Sub(),
	}
	token, err := db.NewAuthToken(ctx, accessTokenTtl, attrs)
	if err != nil {
		return "", err
	}

	key, err := findEcKey(ctx)
	if err != nil {
		return "", err
	}

	claims := pb.Claims{
		Aud: &attrs.Aud,
		Iss: &attrs.ProviderId,
		Sub: &attrs.Sub,
	}

	b, err := proto.Marshal(&claims)
	if err != nil {
		return "", err
	}

	b = append(b, token.Id()...)
	sig, err := key.Sign(ctx, b)
	if err != nil {
		return "", err
	}

	pbToken := pb.Token{
		Claims:    &claims,
		Id:        token.Id(),
		Kid:       key.Kid(),
		Signature: sig,
	}

	b, err = proto.Marshal(&pbToken)
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
		Aud: code.ClientId(),
		Exp: now.Add(idTokenTtl).Unix(),
		Iat: now.Unix(),
		Sub: code.Sub(),
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

	sig, err := key.Sign(ctx, []byte(input))
	if err != nil {
		return "", err
	}

	jwsSig := base64.RawURLEncoding.EncodeToString(sig)

	return fmt.Sprintf("%s.%s", input, jwsSig), nil
}
