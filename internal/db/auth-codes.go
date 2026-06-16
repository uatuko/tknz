package db

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const (
	authCodePrefix = "auth.code|"
)

type AuthCode struct {
	token
}

func (c *AuthCode) Attrs() AuthCodeAttrs {
	return c.attrs.(AuthCodeAttrs)
}

func (c *AuthCode) Id() string {
	return strings.TrimPrefix(c.id, authCodePrefix)
}

type AuthCodeAttrs struct {
	ClientId    string `json:"client_id,omitempty"`
	ProviderId  string `json:"provider_id,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`
	Sub         string `json:"sub,omitempty"`
}

func NewAuthCode(ctx context.Context, ttl time.Duration, attrs AuthCodeAttrs) (*AuthCode, error) {
	code := &AuthCode{
		token: token{
			attrs: attrs,
			exp:   time.Now().Add(ttl).UTC(),
			id:    authCodePrefix + rands(26),
		},
	}

	if err := code.insert(ctx); err != nil {
		return nil, err
	}

	return code, nil
}

func RetrieveAuthCode(ctx context.Context, id string) (*AuthCode, error) {
	code := &AuthCode{}
	if err := retrieveToken(ctx, authCodePrefix+id, &code.token); err != nil {
		return nil, err
	}

	if code.token.attrs != nil {
		var attrs AuthCodeAttrs
		if err := json.Unmarshal(code.token.attrs.([]byte), &attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		code.attrs = attrs
	}

	return code, nil
}
