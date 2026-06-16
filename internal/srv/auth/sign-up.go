package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/felk-ai/idaas/internal/db"
	"github.com/felk-ai/idaas/internal/mail"
	"github.com/felk-ai/idaas/internal/srv/common"
	"github.com/felk-ai/idaas/internal/valid"
	"github.com/rs/zerolog"
)

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		signUpGetHandler(w, r)
	case http.MethodPost:
		signUpPostHandler(w, r)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func signUpGetHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue(codeKey)
	log := zerolog.Ctx(r.Context()).With().Str("otp", code).Logger()
	if code == "" {
		log.Debug().Msg("empty auth otp")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	deleteNonceCookie(w)

	otp, err := db.RetrieveAuthOtp(r.Context(), code)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve otp")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if otp.Expired() {
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	nonce, err := r.Cookie(nonceCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to read nonce cookie")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if err = nonce.Valid(); err != nil {
		log.Info().Err(err).Msg("invalid nonce cookie")
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	h := sha256.Sum224([]byte(otp.Code()))
	if nonce.Value != base64.RawURLEncoding.EncodeToString(h[:]) {
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	if err = otp.Expire(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to expire otp")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	state, err := db.RetrieveAuthState(r.Context(), otp.State())
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if state.Expired() {
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	app, err := db.RetrieveApp(r.Context(), state.AppId())
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve app")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
	}

	attrs := state.Attrs()
	attrs.Login = otp.Login()
	attrs.ProviderId = otp.ProviderId()
	attrs.RpId = app.RpId()
	attrs.RpName = app.RpName()
	attrs.WebAuthnChallenge = db.RandB64Url(webauthnChallengeSize)
	attrs.WebAuthnUserId = db.RandB64Url(webauthnUserIdSize)
	if err = state.SetAttrs(r.Context(), attrs); err != nil {
		log.Error().Err(err).Msg("failed to update state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	u, err := url.Parse(appSignUpPasskeyPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse sign-up passkey path")
		writeError(r.Context(), w, ErrServerError)
		return
	}

	q := u.Query()
	q.Set(challengeKey, state.WebAuthnChallenge())
	q.Set(loginHintKey, state.Login())
	q.Set(rpIdKey, app.RpId())
	q.Set(rpNameKey, app.RpName())
	q.Set(stateKey, state.Id())
	q.Set(uidKey, state.WebAuthnUserId())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}

func signUpPostHandler(w http.ResponseWriter, r *http.Request) {
	stateId := r.FormValue(stateKey)
	log := zerolog.Ctx(r.Context()).With().Str(stateKey, stateId).Logger()
	if stateId == "" {
		log.Debug().Msg("empty auth state")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	if r.FormValue(providerIdKey) == "" {
		log.Debug().Msg("empty provider id")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	login := r.FormValue(loginHintKey)
	if !valid.Email(login) {
		writePasskeySignUpError(w, r, ErrInvalidLogin)
		return
	}

	p, err := db.RetrieveProvider(r.Context(), r.FormValue(providerIdKey))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve provider")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if p.Slug() != db.ProviderSlugPasskey {
		log.Warn().
			Str(providerIdKey, p.Id()).
			Str(slugKey, string(p.Slug())).
			Msg("sign-up attempt blocked")

		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	state, err := db.RetrieveAuthState(r.Context(), stateId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if state.Expired() {
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	otp, err := db.NewAuthOtp(r.Context(), otpTtl, db.AuthOtpAttrs{
		Login:      login,
		ProviderId: p.Id(),
		State:      state.Id(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create otp")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	err = mail.SendSignUpVerify(r.Context(), login, signUpVerifyUri(otp), otpTtl)
	if err != nil {
		log.Error().Err(err).Msg("failed to send sign-up verify email")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	// FIXME: use an encrypted version (instead of the hash)
	h := sha256.Sum224([]byte(otp.Code()))
	setNonceCookie(w, base64.RawURLEncoding.EncodeToString(h[:]))
	writeSignUpVerifyEmail(r.Context(), w, login)
}

func signUpVerifyUri(otp *db.AuthOtp) string {
	u, err := url.JoinPath(common.AuthBaseUrl(), signUpEndpoint)
	if err != nil {
		panic(err)
	}

	q := make(url.Values)
	q.Set(codeKey, otp.Code())

	return fmt.Sprintf("%s?%s", u, q.Encode())
}

func writePasskeySignUpError(w http.ResponseWriter, r *http.Request, code errorCode) {
	u, err := url.Parse(appSignUpErrorPath)
	if err != nil {
		log := zerolog.Ctx(r.Context())

		log.Error().Err(err).Msg("failed to parse url")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	q := u.Query()
	q.Set(providerIdKey, r.FormValue(providerIdKey))
	q.Set(stateKey, r.FormValue(stateKey))
	q.Set(errorCodeKey, code.String())
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}

func writeSignUpVerifyEmail(ctx context.Context, w http.ResponseWriter, login string) {
	data := struct {
		Email      string
		TtlMinutes float64
	}{
		Email:      login,
		TtlMinutes: otpTtl.Minutes(),
	}

	log := zerolog.Ctx(ctx)
	tmpl := signUpVerifyEmailTemplate()
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Error().
			Err(err).
			Str("name", tmpl.Name()).
			Msg("failed to execute template")

		writeError(ctx, w, ErrServerError)
		return
	}
}
