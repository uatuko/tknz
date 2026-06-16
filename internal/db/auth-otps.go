package db

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const (
	authOtpPrefix = "auth.otp|"
)

type AuthOtp struct {
	token
}

func (p *AuthOtp) Attrs() AuthOtpAttrs {
	return p.attrs.(AuthOtpAttrs)
}

func (p *AuthOtp) Code() string {
	return strings.TrimPrefix(p.id, authOtpPrefix)
}

func (p *AuthOtp) Login() string {
	return p.attrs.(AuthOtpAttrs).Login
}

func (p *AuthOtp) State() string {
	return p.attrs.(AuthOtpAttrs).State
}

func (p *AuthOtp) ProviderId() string {
	return p.attrs.(AuthOtpAttrs).ProviderId
}

type AuthOtpAttrs struct {
	Login      string `json:"login,omitempty"`
	ProviderId string `json:"provider_id,omitempty"`
	State      string `json:"state,omitempty"`
}

func NewAuthOtp(ctx context.Context, ttl time.Duration, attrs AuthOtpAttrs) (*AuthOtp, error) {
	otp := &AuthOtp{
		token: token{
			attrs: attrs,
			exp:   time.Now().Add(ttl).UTC(),
			id:    authOtpPrefix + rands(26),
		},
	}

	if err := otp.insert(ctx); err != nil {
		return nil, err
	}

	return otp, nil
}

func RetrieveAuthOtp(ctx context.Context, id string) (*AuthOtp, error) {
	otp := &AuthOtp{}

	err := retrieveToken(ctx, authOtpPrefix+id, &otp.token)
	if err != nil {
		return nil, err
	}

	if otp.token.attrs != nil {
		var attrs AuthOtpAttrs
		if err = json.Unmarshal(otp.token.attrs.([]byte), &attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		otp.attrs = attrs
	}

	return otp, nil
}
