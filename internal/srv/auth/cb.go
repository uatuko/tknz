package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/felk-ai/idaas/internal/db"
	"github.com/felk-ai/idaas/internal/srv/common"
	"github.com/rs/zerolog"
)

func cbHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(r.Context(), w, ErrMethodNotAllowed)
		return
	}

	stateId := r.FormValue(stateKey)
	log := zerolog.Ctx(r.Context()).With().Str(stateKey, stateId).Logger()
	if stateId == "" {
		log.Debug().Msg("empty auth state")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	state, err := db.RetrieveAuthState(r.Context(), stateId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve auth state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if state.Expired() {
		log.Debug().Msg("expired state")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	providerId := state.ProviderId()
	if providerId == "" {
		log.Debug().Msg("empty provider id")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if providerId == canaryProviderId {
		writePasswordCbError(r.Context(), w, ErrInvalidCredentials, stateId, r.FormValue(loginHintKey))
		return
	}

	p, err := db.RetrieveProvider(r.Context(), providerId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve provider")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if p.AppId() != state.AppId() {
		log.Warn().
			Str("app_id_provider", p.AppId()).
			Str("app_id_state", state.AppId()).
			Msg("app id mismatch")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	switch p.Slug() {
	case db.ProviderSlugPasskey:
		cbPasskey(w, r, state, p)
		return

	case db.ProviderSlugPassword:
		cbPassword(w, r, state, p)
		return

	default:
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}
}

func cbPasskey(w http.ResponseWriter, r *http.Request, state *db.AuthState, p *db.Provider) {
	log := zerolog.Ctx(r.Context()).With().Str(stateKey, state.Id()).Logger()

	credential := r.FormValue(credentialKey)
	if credential == "" {
		log.Debug().Msg("empty passkey credential")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	// WebAuthn § 7.2
	// Ref: https://w3c.github.io/webauthn/#sctn-verifying-assertion (Level 3)

	var cred publicKeyCredential
	if err := json.Unmarshal([]byte(credential), &cred); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal public key credential")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if cred.Type != credentialTypePublicKey {
		log.Warn().Str("type", cred.Type).Msg("invalid credential type")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Step 6: user was identified before the authentication ceremony
	idnSrc, err := db.RetrieveIdnSrc(r.Context(), p.Id(), cred.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve passkey identity source")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	// Steps 7-14: client data
	var cData clientData
	if err := json.Unmarshal(cred.Response.ClientDataJSON, &cData); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal public key credential client data")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if cData.Type != clientDataTypeWebauthnGet {
		log.Warn().Str("type", cData.Type).Msg("invalid client data type")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if cData.Challenge != state.WebAuthnChallenge() {
		log.Warn().Str("challenge", cData.Challenge).Msg("client data challenge mismatch")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if !strings.HasPrefix(common.AuthBaseUrl(), cData.Origin) {
		log.Warn().Str("origin", cData.Origin).Msg("unexpected client data origin")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if cData.CrossOrigin {
		log.Warn().Str("topOrigin", cData.TopOrigin).Msg("client data from cross origin")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Authenticator data
	authData := decodeAuthenticatorData(cred.Response.AuthenticatorData)
	if authData == nil {
		log.Warn().Msg("failed to decode authenticator data")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	cr := idnSrc.Cr()

	// Step 15: verify rpIdHash
	rpIdHash := sha256.Sum256([]byte(cr.RpId))
	if rpIdHash != authData.RpIdHash {
		log.Warn().Msg("rp id hash mismatch")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Steps 16-17: UP and UV flags
	if (authData.Flags & authDataFlagUP) != authDataFlagUP {
		log.Warn().Msg("user not present")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if (authData.Flags & authDataFlagUV) != authDataFlagUV {
		log.Warn().Msg("user not verified")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Steps 18-19: BE and BS flags (ignore)

	// Steps 20-21: verify signature
	var cpk credentialPublicKey
	if err = cborDecoder().Unmarshal(cr.PublicKey, &cpk); err != nil {
		log.Error().Err(err).Msg("failed to decode identity source public key")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	hash := sha256.Sum256(cred.Response.ClientDataJSON)
	if err = verifyAssertionSignature(
		&cpk,
		cred.Response.Signature,
		cred.Response.AuthenticatorData,
		hash[:],
	); err != nil {
		log.Warn().Err(err).Msg("failed to verify assertion signature")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	// Step 22: sign count — detect cloned authenticators
	count := cr.SignCount
	if authData.SignCount > 0 || count > 0 {
		if authData.SignCount <= count {
			log.Warn().
				Uint32("have", count).
				Uint32("got", authData.SignCount).
				Msg("sign count mismatch")
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}
	}

	// Step 23: client extensions output (ignore)

	// Step 24: update credential record
	cr.SignCount = authData.SignCount
	cr.BackupState = (authData.Flags & authDataFlagBS) == authDataFlagBS
	if err = idnSrc.SetAttrsCr(r.Context(), cr); err != nil {
		log.Error().Err(err).Msg("failed to update identity source")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	// Step 25: successful authentication ceremony, continue

	if err := state.Expire(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to expire auth state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	code, err := db.NewAuthCode(r.Context(), 5*time.Minute, db.AuthCodeAttrs{
		ClientId:    state.ClientId(),
		ProviderId:  p.Id(),
		RedirectUri: state.RedirectUri(),
		Sub:         idnSrc.IdnId(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth code")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	u, err := url.Parse(state.RedirectUri())
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse redirect uri")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	q := u.Query()
	q.Set(codeKey, code.Id())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}

func cbPassword(w http.ResponseWriter, r *http.Request, state *db.AuthState, p *db.Provider) {
	log := zerolog.Ctx(r.Context()).With().Str(stateKey, state.Id()).Logger()

	sub := r.FormValue(loginHintKey)
	password := r.FormValue(passwordKey)
	if sub == "" || password == "" {
		log.Debug().Msg("empty subject or password")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	if p.AppId() != state.Attrs().AppId {
		log.Warn().
			Str("app_id_provider", p.AppId()).
			Str("app_id_state", state.Attrs().AppId).
			Msg("app id in auth state and authentication provider doesn't match")

		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if p.Slug() != db.ProviderSlugPassword {
		log.Warn().
			Str("slug", string(p.Slug())).
			Str("provider_id", p.Id()).
			Msg("unexpected sign-in request")

		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	idnSrc, err := db.RetrieveIdnSrc(r.Context(), p.Id(), sub)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writePasswordCbError(r.Context(), w, ErrInvalidCredentials, state.Id(), sub)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve identity source")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if err := verifyCredentials(password, idnSrc.Pwd()); err != nil {
		writePasswordCbError(r.Context(), w, ErrInvalidCredentials, state.Id(), sub)
		return
	}

	if err := state.Expire(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to expire auth state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	code, err := db.NewAuthCode(r.Context(), 5*time.Minute, db.AuthCodeAttrs{
		ClientId:    state.ClientId(),
		ProviderId:  p.Id(),
		RedirectUri: state.RedirectUri(),
		Sub:         idnSrc.IdnId(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth code")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	u, err := url.Parse(state.RedirectUri())
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse redirect uri")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	q := u.Query()
	q.Set(codeKey, code.Id())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}

func writePasswordCbError(ctx context.Context, w http.ResponseWriter, code errorCode, stateId string, loginHint string) {
	u, err := url.Parse(appSignInPasswordPath)
	if err != nil {
		log := zerolog.Ctx(ctx).With().Str(stateKey, stateId).Logger()

		log.Error().Err(err).Msg("failed to parse url")
		writeError(ctx, w, ErrServerError)
		return
	}

	q := u.Query()
	q.Set(errorCodeKey, code.String())
	q.Set(loginHintKey, loginHint)
	q.Set(stateKey, stateId)
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}
