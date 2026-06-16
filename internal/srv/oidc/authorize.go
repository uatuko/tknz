package oidc

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/rs/zerolog"

	"github.com/felk-ai/idaas/internal/db"
	"github.com/felk-ai/idaas/internal/srv/auth"
)

func AuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", fmt.Sprintf("%v, %v", http.MethodGet, http.MethodPost))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// TODO: authentication request validation
	// Ref: https://openid.net/specs/openid-connect-core-1_0.html#AuthRequestValidation

	clientId := r.FormValue(clientIdKey)
	log := zerolog.Ctx(r.Context()).With().Str(clientIdKey, clientId).Logger()
	if clientId == "" {
		log.Debug().Msg("empty client id")
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	app, err := db.RetrieveAppByOAuthClientId(r.Context(), clientId)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(r.Context(), w, ErrInvalidRequest)
			return
		}

		log.Error().Err(err).Msg("failed to retrieve app")
		writeError(r.Context(), w, ErrTemporarilyUnavailable)
		return
	}

	if r.FormValue(responseTypeKey) != responseTypeCode {
		writeError(r.Context(), w, ErrUnsupportedResponseType)
		return
	}

	redirectUri := r.FormValue(redirectUriKey)
	if redirectUri == "" {
		writeError(r.Context(), w, ErrInvalidRequest)
		return
	}

	if !slices.Contains(app.RedirectUris(), redirectUri) {
		writeError(r.Context(), w, ErrUnauthorizedRedirectUri)
		return
	}

	auth.WriteAuthRedirect(r.Context(), w, app, redirectUri, r.FormValue(loginHintKey))
}
