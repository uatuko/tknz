package db

import (
	"context"
	"strings"
	"time"
)

const (
	authTokenPrefix = "auth.token|"
)

type AuthToken struct {
	token
}

func (t *AuthToken) Attrs() AuthTokenAttrs {
	return t.attrs.(AuthTokenAttrs)
}

func (t *AuthToken) Token() []byte {
	str := strings.TrimPrefix(t.id, authTokenPrefix)
	b, err := encoding.DecodeString(str)
	if err != nil {
		panic(err)
	}

	return b
}

type AuthTokenAttrs struct {
	AppId      string `json:"app_id,omitempty"`
	ProviderId string `json:"provider_id,omitempty"`

	Aud string `json:"aud,omitempty"`
	Sub string `json:"sub,omitempty"`
}

func NewAuthToken(ctx context.Context, ttl time.Duration, attrs AuthTokenAttrs) (*AuthToken, error) {
	token := &AuthToken{
		token: token{
			attrs: attrs,
			exp:   time.Now().Add(ttl).UTC(),
			id:    authTokenPrefix + rands(32),
		},
	}

	if err := token.insert(ctx); err != nil {
		return nil, err
	}

	return token, nil
}
