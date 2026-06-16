package jwt

import (
	"fmt"
	"time"
)

type Claims struct {
	public registeredClaims
}

func (c *Claims) Aud() string {
	return c.public.Aud
}

func (c *Claims) Exp() time.Time {
	return time.Unix(c.public.Exp, 0)
}

func (c *Claims) Iat() time.Time {
	return time.Unix(c.public.Iat, 0)
}

func (c *Claims) Iss() string {
	return c.public.Iss
}

func (c *Claims) Jti() string {
	return c.public.Jti
}

func (c *Claims) Nbf() time.Time {
	return time.Unix(c.public.Nbf, 0)
}

func (c *Claims) Private() map[string]any {
	return nil
}

func (c *Claims) Sub() string {
	return c.public.Sub
}

// Ref: https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1 (registered claims)
type registeredClaims struct {
	Aud string `json:"aud,omitempty"`
	Exp int64  `json:"exp,omitempty"`
	Iat int64  `json:"iat,omitempty"`
	Iss string `json:"iss,omitempty"`
	Jti string `json:"jti,omitempty"`
	Nbf int64  `json:"nbf,omitempty"`
	Sub string `json:"sub,omitempty"`
}

func (c *registeredClaims) validate() error {
	if c.Exp < 0 || c.Iat < 0 || c.Exp < c.Iat {
		return fmt.Errorf("invalid jwt")
	}

	t := time.Now().UTC().Unix()
	if c.Exp < t {
		return fmt.Errorf("jwt expired")
	}

	if c.Iat > t {
		return fmt.Errorf("jwt from future")
	}

	if c.Nbf > 0 && c.Nbf > t {
		return fmt.Errorf("jwt not valid yet")
	}

	return nil
}
