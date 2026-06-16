package common

import (
	"net/url"
	"os"
	"strings"
)

const (
	SysSpaceId = "sys"

	authPath = "/auth/"
	oidcPath = "/oidc/"
)

func AuthBaseUrl() string {
	s, err := url.JoinPath(os.Getenv("BASE_URL"), authPath)
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

func OidcPathPattern() string {
	return oidcPath
}

func OidcPathPrefix() string {
	return strings.TrimSuffix(oidcPath, "/")
}
