package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"

	"github.com/rs/zerolog"
	"go.tknz.dev/internal/db"
)

type passkeySignInData struct {
	Challenge     []byte
	CredentialIds [][]byte

	_ struct{} `cbor:",toarray"`
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
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

	login := r.FormValue(loginHintKey)
	if login == "" {
		log.Debug().Msg("empty login hint")
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

	providers, err := db.ListProvidersByLogin(r.Context(), state.AppId(), login)
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve providers")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if len(providers) == 0 {
		// No providers linked to login (or login doesn't exist)
		// Fallback to password provider
		p, err := db.RetrieveProviderBySlub(r.Context(), state.AppId(), db.ProviderSlugPassword)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(r.Context(), w, ErrAccessDenied)
				return
			}

			log.Error().Err(err).Msg("failed to retrieve password provider")
			writeError(r.Context(), w, ErrTemporarilyUnavailable)
			return
		}

		writePasswordSignInRedirect(r.Context(), w, state, p.Id(), login)
		return
	}

	for _, p := range providers {
		switch p.Slug() {
		case db.ProviderSlugPassword:
			writePasswordSignInRedirect(r.Context(), w, state, p.Id(), login)
			return

		case db.ProviderSlugPasskey:
			writePasskeySignInRedirect(r.Context(), w, state, p.Id(), login)
			return
		}
	}

	// No useable provider, use a canary
	writePasswordSignInRedirect(r.Context(), w, state, canaryProviderId, login)
}

func writePasskeySignInRedirect(ctx context.Context, w http.ResponseWriter, state *db.AuthState, providerId string, login string) {
	log := zerolog.Ctx(ctx).With().Str(stateKey, state.Id()).Logger()

	challenge := make([]byte, webauthnChallengeSize)
	rand.Read(challenge)

	attrs := state.Attrs()
	attrs.Login = login
	attrs.ProviderId = providerId
	attrs.WebAuthnChallenge = base64.RawURLEncoding.EncodeToString(challenge)
	if err := state.SetAttrs(ctx, attrs); err != nil {
		log.Error().Err(err).Msg("failed to update state")
		writeError(ctx, w, ErrTemporarilyUnavailable)
		return
	}

	srcs, err := db.ListIdnSrcByProviderIdAndLogin(ctx, state.ProviderId(), login)
	if err != nil {
		log.Error().Err(err).Msg("failed to list identity sources")
		writeError(ctx, w, ErrTemporarilyUnavailable)
		return
	}

	credIds := make([][]byte, 0, len(srcs))
	for _, s := range srcs {
		id, err := base64.RawURLEncoding.DecodeString(s.Sub())
		if err != nil {
			log.Error().Err(err).Msg("failed to decode credential id")
			writeError(ctx, w, ErrServerError)
			return
		}

		credIds = append(credIds, id)
	}

	data, err := cborEncoder().Marshal(passkeySignInData{
		Challenge:     challenge,
		CredentialIds: credIds,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal public key sign-in data")
		writeError(ctx, w, ErrServerError)
		return
	}

	u, err := url.Parse(appSignInPasskeyPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse url")
		writeError(ctx, w, ErrServerError)
		return
	}

	q := u.Query()
	q.Set(dataKey, base64.RawURLEncoding.EncodeToString(data))
	q.Set(stateKey, state.Id())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}

func writePasswordSignInRedirect(ctx context.Context, w http.ResponseWriter, state *db.AuthState, providerId string, loginHint string) {
	log := zerolog.Ctx(ctx).With().Str(stateKey, state.Id()).Logger()

	if providerId != "" {
		attrs := state.Attrs()
		attrs.ProviderId = providerId

		if err := state.SetAttrs(ctx, attrs); err != nil {
			log.Error().Err(err).Msg("failed to update state")
			writeError(ctx, w, ErrTemporarilyUnavailable)
			return
		}
	}

	u, err := url.Parse(appSignInPasswordPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse url")
		writeError(ctx, w, ErrServerError)
		return
	}

	q := u.Query()
	q.Set(loginHintKey, loginHint)
	q.Set(stateKey, state.Id())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}
