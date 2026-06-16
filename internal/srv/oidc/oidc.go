package oidc

import (
	"net/http"
)

const (
	AuthorizationEndpoint = "/authorize"
	TokenEndpoint         = "/token"
	UserinfoEndpoint      = "/userinfo"

	DiscoveryEndpoint = "/.well-known/openid-configuration"
	JwksEndpoint      = "/.well-known/jwks.json"

	clientAssertionKey     = "client_assertion"
	clientAssertionTypeKey = "client_assertion_type"
	clientIdKey            = "client_id"
	codeKey                = "code"
	grantTypeKey           = "grant_type"
	loginHintKey           = "login_hint"
	redirectUriKey         = "redirect_uri"
	responseTypeKey        = "response_type"

	clientAssertionTypeJwtBearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

	grantTypeAuthorizationCode = "authorization_code"

	responseTypeCode = "code"

	tokenTypeBearer = "Bearer"

	joseHeaderAlgES256 = "ES256"
	joseHeaderAlgNone  = "none"
	joseHeaderTypJwt   = "JWT"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(AuthorizationEndpoint, AuthorizationHandler)
	mux.HandleFunc(TokenEndpoint, TokenHandler)
	mux.HandleFunc(JwksEndpoint, JwksHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(r.Context(), w, ErrNotFound)
	})

	return mux
}
