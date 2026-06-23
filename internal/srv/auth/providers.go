package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/kms"
	"go.tknz.dev/internal/srv/common"
)

type oidcProvider struct {
	ClientId    string `json:"client_id,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`

	provider
}

type provider struct {
	Id   string          `json:"id"`
	Slug db.ProviderSlug `json:"slug"`
}

type providersResponse struct {
	Oidc     []oidcProvider `json:"oidc,omitempty"`
	SignUp   *provider      `json:"sign_up,omitempty"`
	UseLogin bool           `json:"use_login,omitempty"`
}

func providersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJsonError(r.Context(), w, ErrMethodNotAllowed)
		return
	}

	stateId := r.FormValue(stateKey)
	log := zerolog.Ctx(r.Context()).With().Str(stateKey, stateId).Logger()
	if stateId == "" {
		log.Debug().Msg("empty auth state")
		writeJsonError(r.Context(), w, ErrInvalidRequest)
		return
	}

	state, err := db.RetrieveAuthState(r.Context(), stateId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJsonError(r.Context(), w, ErrAccessDenied)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve auth state")
		writeJsonError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if state.Exp().Before(time.Now()) {
		writeJsonError(r.Context(), w, ErrAccessDenied)
		return
	}

	providers, err := db.ListProviders(r.Context(), state.AppId())
	if err != nil {
		log.Error().Err(err).Msg("failed to list providers")
		writeJsonError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	resp := providersResponse{}
	for _, p := range providers {
		switch p.Slug() {
		case db.ProviderSlugGoogleOAuth:
			if resp.Oidc == nil {
				resp.Oidc = make([]oidcProvider, 0, 1)
			}

			resp.Oidc = append(resp.Oidc, oidcProvider{
				ClientId:    p.ClientId(),
				RedirectUri: redirectUri(&p),

				provider: provider{
					Id:   p.Id(),
					Slug: p.Slug(),
				},
			})

		case db.ProviderSlugPasskey:
			resp.SignUp = &provider{
				Id:   p.Id(),
				Slug: p.Slug(),
			}

			resp.UseLogin = true

		case db.ProviderSlugPassword:
			resp.UseLogin = true
		}
	}

	b, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal response")
		writeJsonError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	w.Header().Add("Cache-Control", "private")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func providersCbHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
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
		writeError(r.Context(), w, ErrAccessDenied)
		return
	}

	p, err := db.RetrieveProvider(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrInvalidRequest)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve provider")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if p.AppId() != state.Attrs().AppId {
		log.Warn().
			Str("app_id_provider", p.AppId()).
			Str("app_id_state", state.Attrs().AppId).
			Msg("app id in auth state and authentication provider doesn't match")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	clientSecret, err := kms.Decrypt(r.Context(), p.ClientSeceret())
	if err != nil {
		log.Error().Err(err).Msg("failed to decrypt client secret")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if err := state.Expire(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to expire auth state")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	payload := url.Values{}
	payload.Add(grantTypeKey, grantTypeAuthorizationCode)
	payload.Add(codeKey, r.FormValue(codeKey))
	payload.Add(clientIdKey, p.ClientId())
	payload.Add(clientSecretKey, string(clientSecret))
	payload.Add(redirectUriKey, redirectUri(p))

	// TODO: handle based on provider
	resp, err := http.PostForm(googleOAuthTokenEndpoint, payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to authentication provider")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read response")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status_code", resp.StatusCode).Msg("unexpected response status")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	// TODO: response validation (including 'id_token' validation)
	var tokenResp struct {
		IdToken string `json:"id_token"`
	}
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		log.Warn().Err(err).Msg("failed to parse token response")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	chunks := strings.Split(tokenResp.IdToken, ".")
	if len(chunks) != 3 {
		log.Warn().Err(err).Msg("invalid id token")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	idTokenBytes := make([]byte, base64.RawURLEncoding.DecodedLen(len(chunks[1])))
	if _, err := base64.RawURLEncoding.Decode(idTokenBytes, []byte(chunks[1])); err != nil {
		log.Warn().Err(err).Msg("failed to decode id token")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	var idToken struct {
		Aud   string
		Email string
		Sub   string
	}

	if err := json.Unmarshal(idTokenBytes, &idToken); err != nil {
		log.Warn().Err(err).Msg("failed to parse id token")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	login := idToken.Email
	if login == "" {
		login = idToken.Sub
	}

	idn, err := db.RetrieveIdnBySrc(r.Context(), p.Id(), idToken.Sub)
	// TODO: consider "syncing" provider attrs to db (e.g. picture, name, email etc.)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			// FIXME: it's possible we already have the same app+login from a different provider
			// TODO: control sign-ups (i.e. don't create a new identity if sign-ups aren't allowed)
			idn, err = db.NewIdn(r.Context(), p.AppId(), login, db.IdnAttrs{
				Email: idToken.Email,
			})

			if err != nil {
				log.Error().Err(err).Msg("failed to store identity")
				writeError(r.Context(), w, ErrTemporarilyUnavailable)
				return
			}

			_, err := db.NewIdnSrc(r.Context(), idn.Id(), p.Id(), idToken.Sub)
			if err != nil {
				log.Error().Err(err).Msg("failed to store identity source")
				writeError(r.Context(), w, ErrTemporarilyUnavailable)
				return
			}
		} else {
			log.Error().Err(err).Msg("failed to retrieve identity")
			writeError(r.Context(), w, ErrTemporarilyUnavailable)
			return
		}
	}

	code, err := db.NewAuthCode(r.Context(), 5*time.Minute, db.AuthCodeAttrs{
		ClientId:    state.ClientId(),
		ProviderId:  p.Id(),
		RedirectUri: state.RedirectUri(),
		Sub:         idn.Id(),
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to generate code")
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

func redirectUri(p *db.Provider) string {
	if p.Slug() != db.ProviderSlugGoogleOAuth {
		return ""
	}

	s, err := url.JoinPath(common.AuthBaseUrl(),
		strings.Replace(providersCbEndpoint, "{id}", p.Id(), 1))
	if err != nil {
		panic(err)
	}

	return s
}
