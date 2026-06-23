package auth

import (
	"net/http"
	"strings"

	"go.tknz.dev/internal/srv/common"
)

func deleteNonceCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: nonceCookieName,

		HttpOnly: true,
		MaxAge:   -1,
		Path:     common.AuthPathPattern(),
		SameSite: http.SameSiteStrictMode,
		Secure:   !strings.HasPrefix("http://", common.AuthBaseUrl()),
	})
}

func setNonceCookie(w http.ResponseWriter, nonce string) {
	http.SetCookie(w, &http.Cookie{
		Name:  nonceCookieName,
		Value: nonce,

		HttpOnly: true,
		MaxAge:   int(otpTtl.Seconds()),
		Path:     common.AuthPathPattern(),
		SameSite: http.SameSiteStrictMode,
		Secure:   !strings.HasPrefix("http://", common.AuthBaseUrl()),
	})
}
