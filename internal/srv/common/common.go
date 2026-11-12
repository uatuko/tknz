package common

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	SysSpaceId = "sys"

	authPath = "/auth/"
	oidcPath = "/oidc/"
)

var baseUrl string

func AuthBaseUrl() string {
	s, err := url.JoinPath(baseUrl, authPath)
	if err != nil {
		panic(err)
	}

	return s
}

func AuthPathPattern() string {
	return authPath
}

func AuthPathPrefix() string {
	return strings.TrimSuffix(authPath, "/")
}

func OidcBaseUrl() string {
	s, err := url.JoinPath(baseUrl, oidcPath)
	if err != nil {
		panic(err)
	}

	return s
}

func OidcPathPattern() string {
	return oidcPath
}

func OidcPathPrefix() string {
	return strings.TrimSuffix(oidcPath, "/")
}

// SetBaseUrl sets the base url (e.g. https://app.tknz.dev) used to resolve relative urls.
func SetBaseUrl(s string) error {
	if s == "" {
		return fmt.Errorf("missing base url")
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid base url, err: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid base url scheme (want: 'http' or 'https', have: '%v')", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("invalid base url, missing host")
	}

	baseUrl = s
	return nil
}
