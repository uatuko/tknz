package db

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	authStatePrefix = "auth.state|"
)

type AuthState struct {
	token
}

func (s *AuthState) AppId() string {
	return s.attrs.(AuthStateAttrs).AppId
}

func (s *AuthState) Attrs() AuthStateAttrs {
	return s.attrs.(AuthStateAttrs)
}

func (s *AuthState) ClientId() string {
	return s.attrs.(AuthStateAttrs).ClientId
}

func (s *AuthState) Id() string {
	return strings.TrimPrefix(s.id, authStatePrefix)
}

func (s *AuthState) Login() string {
	return s.attrs.(AuthStateAttrs).Login
}

func (s *AuthState) ProviderId() string {
	return s.attrs.(AuthStateAttrs).ProviderId
}

func (s *AuthState) RedirectUri() string {
	return s.attrs.(AuthStateAttrs).RedirectUri
}

func (s *AuthState) Rev() int32 {
	return s.attrs.(AuthStateAttrs).Rev
}

func (s *AuthState) RpId() string {
	return s.attrs.(AuthStateAttrs).RpId
}

func (s *AuthState) RpName() string {
	return s.attrs.(AuthStateAttrs).RpName
}

func (s *AuthState) SetAttrs(ctx context.Context, attrs AuthStateAttrs) error {
	attrs.Rev = s.Rev() + 1

	qry := `
		update tokens
		set
			attrs = $3::jsonb
		where
			id = $1::text
			and attrs['_rev']::integer =  $2::integer
		returning attrs['_rev']::integer;
	`

	if err := pg.QueryRow(ctx, qry, s.id, s.Rev(), attrs).Scan(&attrs.Rev); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return wrapError(err, ErrConflict)
		}

		return wrapError(err, ErrUnknown)
	}

	s.attrs = attrs
	return nil
}

func (s *AuthState) WebAuthnChallenge() string {
	return s.attrs.(AuthStateAttrs).WebAuthnChallenge
}

func (s *AuthState) WebAuthnUserId() string {
	return s.attrs.(AuthStateAttrs).WebAuthnUserId
}

type AuthStateAttrs struct {
	AppId       string `json:"app_id,omitempty"`
	ClientId    string `json:"client_id,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`

	Login             string `json:"login,omitempty"`
	ProviderId        string `json:"provider_id,omitempty"`
	RpId              string `json:"rp_id,omitempty"`
	RpName            string `json:"rp_name,omitempty"`
	WebAuthnChallenge string `json:"wa_challenge,omitempty"` // 16 bytes min
	WebAuthnUserId    string `json:"wa_uid,omitempty"`       // 64 bytes max

	Rev int32 `json:"_rev,omitempty"`
}

func NewAuthState(ctx context.Context, ttl time.Duration, attrs AuthStateAttrs) (*AuthState, error) {
	attrs.Rev = rand.Int32()

	state := &AuthState{
		token: token{
			attrs: attrs,
			exp:   time.Now().Add(ttl).UTC(),
			id:    authStatePrefix + rands(26),
		},
	}

	if err := state.insert(ctx); err != nil {
		return nil, err
	}

	return state, nil
}

func RetrieveAuthState(ctx context.Context, id string) (*AuthState, error) {
	state := &AuthState{}
	if err := retrieveToken(ctx, authStatePrefix+id, &state.token); err != nil {
		return nil, err
	}

	if state.token.attrs != nil {
		var attrs AuthStateAttrs
		if err := json.Unmarshal(state.token.attrs.([]byte), &attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		state.attrs = attrs
	}

	return state, nil
}
