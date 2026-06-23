package oidc

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/srv/common"
)

type jwkEc struct {
	Alg string `json:"alg,omitempty"`
	Kid string `json:"kid,omitempty"`
	Kty string `json:"kty,omitempty"`
	Use string `json:"use,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

type jwkSet struct {
	Keys []jwkEc `json:"keys"`
}

func JwksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log := zerolog.Ctx(r.Context())

	jwks, err := db.ListJwks(r.Context(), common.SysSpaceId)
	if err != nil {
		log.Error().Err(err).Msg("failed to list jwks")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	keys := make([]jwkEc, 0, len(jwks))
	for _, jwk := range jwks {
		if jwk.Params().Kty == db.JwkKtyEc {
			keys = append(keys, jwkEc{
				Alg: string(jwk.Params().Alg),
				Kid: jwk.Kid(),
				Kty: string(jwk.Params().Kty),
				Use: string(jwk.Params().Use),
				Crv: string(jwk.Params().Crv),
				X:   jwk.Params().X,
				Y:   jwk.Params().Y,
			})
		}
	}

	resp := jwkSet{Keys: keys}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal jwks")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
