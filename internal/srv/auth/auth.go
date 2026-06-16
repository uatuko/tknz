package auth

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"go.tknz.dev/internal/db"
)

const (
	AccessTokenPrefix = "9lizNw4."

	appSignInPath         = "/sign-in"
	appSignInPasskeyPath  = "/sign-in/passkey"
	appSignInPasswordPath = "/sign-in/password"
	appSignUpErrorPath    = "/sign-up"
	appSignUpPasskeyPath  = "/sign-up/passkey"

	cbEndpoint          = "/cb"
	credentialsEndpoint = "/credentials"
	providersEndpoint   = "/providers"
	providersCbEndpoint = "/providers/{id}/cb"
	signInEndpoint      = "/sign-in"
	signUpEndpoint      = "/sign-up"

	challengeKey    = "challenge"
	clientIdKey     = "client_id"
	clientSecretKey = "client_secret"
	codeKey         = "code"
	credentialKey   = "credential"
	dataKey         = "data"
	errorCodeKey    = "error_code"
	grantTypeKey    = "grant_type"
	loginHintKey    = "login_hint"
	passwordKey     = "password"
	providerIdKey   = "provider_id"
	redirectUriKey  = "redirect_uri"
	rpIdKey         = "rp_id"
	rpNameKey       = "rp_name"
	slugKey         = "slug"
	stateKey        = "state"
	subjectKey      = "sub"
	uidKey          = "uid"

	canaryProviderId = "canary"

	grantTypeAuthorizationCode = "authorization_code"

	joseHeaderAlgES256 = "ES256"

	nonceCookieName = "_nonce"

	webauthnChallengeSize int = 16
	webauthnUserIdSize    int = 64

	accessTokenTtl = 6 * time.Hour
	otpTtl         = 10 * time.Minute
	idTokenTtl     = 15 * time.Minute
	stateTtl       = 15 * time.Minute
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(cbEndpoint, cbHandler)
	mux.HandleFunc(credentialsEndpoint, credentialsHandler)
	mux.HandleFunc(providersEndpoint, providersHandler)
	mux.HandleFunc(providersCbEndpoint, providersCbHandler)
	mux.HandleFunc(signInEndpoint, signInHandler)
	mux.HandleFunc(signUpEndpoint, signUpHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(r.Context(), w, ErrNotFound)
	})

	return mux
}

func WriteAuthRedirect(ctx context.Context, w http.ResponseWriter, app *db.App, redirectUri string, loginHint string) {
	state, err := db.NewAuthState(ctx, stateTtl, db.AuthStateAttrs{
		AppId:       app.Id(),
		ClientId:    app.OAuthClientId(),
		RedirectUri: redirectUri,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth state")
		writeError(ctx, w, ErrServerError)
		return
	}

	u, err := url.Parse(appSignInPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse url")
		writeError(ctx, w, ErrServerError)
		return
	}

	q := u.Query()
	q.Set(stateKey, state.Id())
	if loginHint != "" {
		q.Set(loginHintKey, loginHint)
	}
	u.RawQuery = q.Encode()

	w.Header().Add("Location", u.String())
	w.WriteHeader(http.StatusFound)
}
