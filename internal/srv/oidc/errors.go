package oidc

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/rs/zerolog"
)

const (
	// OAuth 2.0 errors

	ErrInvalidRequest          errorCode = "invalid_request"
	ErrUnauthorizedRedirectUri errorCode = "unauthorized_redirect_uri"
	ErrUnsupportedResponseType errorCode = "unsupported_response_type"
	ErrServerError             errorCode = "server_error"
	ErrTemporarilyUnavailable  errorCode = "temporarily_unavailable"

	// OP (server) errors

	ErrNotFound errorCode = "not_found"
	ErrUnknown  errorCode = "unknown"

	errorRpTmplFile     = ".dist/tmpl/oidc/errors/rp.html"
	errorServerTmplFile = ".dist/tmpl/oidc/errors/server.html"
)

var (
	errorRpTmpl     *template.Template
	errorServerTmpl *template.Template
)

type errorCode string

func (c errorCode) Error() string {
	switch c {
	case ErrInvalidRequest:
		return "invalid request"
	case ErrNotFound:
		return "resource not found"
	case ErrServerError:
		return "server error"
	case ErrTemporarilyUnavailable:
		return "temporarily unavailable"
	case ErrUnauthorizedRedirectUri:
		return "unauthorized redirect uri"
	case ErrUnsupportedResponseType:
		return "unsupported response type"

	default:
		return "unknown error"
	}
}

func (c errorCode) String() string {
	return string(c)
}

type Error struct {
	ErrorCode        errorCode `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`

	statusCode int
}

func (e Error) Error() string {
	return e.ErrorCode.String()
}

func (e Error) StatusCode() int {
	return e.statusCode
}

func (e Error) tmpl() *template.Template {
	var tmpl *template.Template

	switch e.ErrorCode {
	// Authorization errors

	case ErrInvalidRequest,
		ErrUnauthorizedRedirectUri,
		ErrUnsupportedResponseType:

		if errorRpTmpl == nil {
			errorRpTmpl = template.Must(template.ParseFiles(errorRpTmplFile))
		}
		tmpl = errorRpTmpl

	// Server errors

	default:
		if errorServerTmpl == nil {
			errorServerTmpl = template.Must(template.ParseFiles(errorServerTmplFile))
		}
		tmpl = errorServerTmpl
	}

	return tmpl
}

func NewError(code errorCode) *Error {
	e := &Error{ErrorCode: code}

	switch code {
	// RP (client)

	case ErrInvalidRequest:
		e.statusCode = http.StatusBadRequest
		e.ErrorDescription = "The authorization request is invalid or malformed."

	case ErrUnauthorizedRedirectUri:
		e.statusCode = http.StatusForbidden
		e.ErrorDescription = "Unauthorized redirect URI."

	case ErrUnsupportedResponseType:
		e.statusCode = http.StatusBadRequest
		e.ErrorDescription = "The server does not support the requested authorization processing flow, or it is invalid."

	// OP (server)

	case ErrNotFound:
		e.statusCode = http.StatusNotFound
		e.ErrorDescription = "The resource you are looking for is not found on this server, or you assembled the link incorrectly."

	case ErrServerError:
		e.statusCode = http.StatusInternalServerError
		e.ErrorDescription = "The server encountered an unexpected condition that prevented it from completing the request."

	case ErrTemporarilyUnavailable:
		e.statusCode = http.StatusServiceUnavailable
		e.ErrorDescription = "Service temporarily unavailable."

	default:
		e.statusCode = http.StatusInternalServerError
		e.ErrorDescription = "Unknown error"
	}

	return e
}

func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	log := zerolog.Ctx(ctx)

	var e *Error
	if !errors.As(err, &e) {
		var code errorCode
		if !errors.As(err, &code) {
			log.Warn().
				Err(err).
				Msg("unknown oidc error")

			code = ErrUnknown
		}

		e = NewError(code)
	}

	log.Debug().
		Int("status_code", e.StatusCode()).
		Str("error", e.Error()).
		Str("error_description", e.ErrorDescription).
		Msg("writing oidc error")

	w.WriteHeader(e.statusCode)
	if err := e.tmpl().Execute(w, e); err != nil {
		log.Error().
			Err(err).
			Msg("failed to use error template")

		http.Error(w, e.ErrorDescription, e.statusCode)
	}
}
