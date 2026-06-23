package oidc

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/jose/jws"
	"go.tknz.dev/internal/srv/auth"
)

// Ref: https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1 (registered claims)
type jwtClaims struct {
	Aud string `json:"aud,omitempty"`
	Exp int64  `json:"exp,omitempty"`
	Iat int64  `json:"iat,omitempty"`
	Iss string `json:"iss,omitempty"`
	Jti string `json:"jti,omitempty"`
	Nbf int64  `json:"nbf,omitempty"`
	Sub string `json:"sub,omitempty"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	ExpiresIn    uint16 `json:"expires_in,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// TODO: token request validation
	// Ref: https://openid.net/specs/openid-connect-core-1_0.html#TokenRequestValidation

	clientId := r.FormValue(clientIdKey)
	log := zerolog.Ctx(r.Context()).With().Str(clientIdKey, clientId).Logger()
	if clientId == "" {
		log.Debug().Msg("empty client id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.FormValue(grantTypeKey) != grantTypeAuthorizationCode {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.FormValue(clientAssertionTypeKey) != clientAssertionTypeJwtBearer {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	app, err := db.RetrieveAppByOAuthClientId(r.Context(), clientId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claims := jwtClaims{
		Aud: os.Getenv("OIDC_BASE_URL"),
		Iss: app.OAuthClientId(),
		Sub: app.OAuthClientId(),
	}
	attrs := app.Attrs()
	if err := jwtValidate(r.FormValue(clientAssertionKey), claims, attrs.Keys); err != nil {
		log.Warn().Err(err).Msg("failed to validate client assertion")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code, err := db.RetrieveAuthCode(r.Context(), r.FormValue(codeKey))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if code.Expired() {
		// FIXME: better errors
		w.WriteHeader(http.StatusGone)
		return
	}

	if err := code.Expire(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to expire oidc code")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if app.OAuthClientId() != code.Attrs().ClientId {
		log.Warn().
			Str("client_id_app", app.OAuthClientId()).
			Str("client_id_code", code.Attrs().ClientId).
			Msg("client id for oidc code and app doesn't match")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessToken, err := auth.NewAccessToken(r.Context(), app, code)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate a new access token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	idToken, err := auth.NewIdToken(r.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate a new id token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := tokenResponse{
		AccessToken: accessToken,
		IdToken:     idToken,
		TokenType:   tokenTypeBearer,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func jwsValidate(payload string, b64sig string, key *ecdsa.PublicKey) error {
	sig, err := base64.RawURLEncoding.DecodeString(b64sig)
	if err != nil {
		return err
	}

	digest := sha256.Sum256([]byte(payload))
	var r, s big.Int
	r.SetBytes(sig[:32])
	s.SetBytes(sig[32:])

	if !ecdsa.Verify(key, digest[:], &r, &s) {
		return fmt.Errorf("jws validation failed")
	}

	return nil
}

func jwtValidate(token string, want jwtClaims, keys []db.JwkParams) error {
	chunks := strings.Split(token, ".")
	if len(chunks) != 3 {
		return fmt.Errorf("invalid token")
	}

	b, err := base64.RawURLEncoding.DecodeString(chunks[0])
	if err != nil {
		return err
	}

	var header jws.Header
	err = json.Unmarshal(b, &header)
	if err != nil {
		return err
	}

	if header.Typ != joseHeaderTypJwt {
		return fmt.Errorf("unknown typ: %s", header.Typ)
	}

	if header.Alg == joseHeaderAlgNone {
		if len(chunks[3]) != 0 {
			return fmt.Errorf("invalid signature")
		}

		return jwtValidateClaims(chunks[1], want)
	}

	if header.Alg != joseHeaderAlgES256 {
		return fmt.Errorf("unknown alg: %s", header.Alg)
	}

	var key *ecdsa.PublicKey
	for _, k := range keys {
		if header.Kid == "" || header.Kid == k.Kid {
			key, err = k.EcPublicKey()
			if err != nil {
				return err
			}

			break
		}
	}

	if key == nil {
		return fmt.Errorf("jwk not found")
	}

	err = jwsValidate(fmt.Sprintf("%s.%s", chunks[0], chunks[1]), chunks[2], key)
	if err != nil {
		return err
	}

	return jwtValidateClaims(chunks[1], want)
}

func jwtValidateClaims(claims string, want jwtClaims) error {
	b, err := base64.RawURLEncoding.DecodeString(claims)
	if err != nil {
		return err
	}

	var have jwtClaims
	if err := json.Unmarshal(b, &have); err != nil {
		return err
	}

	if have.Exp < 0 || have.Iat < 0 || have.Exp < have.Iat {
		return fmt.Errorf("invalid jwt")
	}

	t := time.Now().UTC().Unix()
	if have.Exp < t {
		return fmt.Errorf("jwt expired")
	}

	if have.Iat > t {
		return fmt.Errorf("jwt from future")
	}

	if have.Nbf > 0 && have.Nbf > t {
		return fmt.Errorf("jwt not valid yet")
	}

	// TODO: check max ttl (exp - iat < max allowed)

	if want.Aud != "" && want.Aud != have.Aud {
		return fmt.Errorf("invalid aud")
	}

	if want.Iss != "" && want.Iss != have.Iss {
		return fmt.Errorf("invalid iss")
	}

	if want.Sub != "" && want.Sub != have.Sub {
		return fmt.Errorf("invlid sub")
	}

	return nil
}
