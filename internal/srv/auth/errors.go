package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"
)

const (
	ErrAccessDenied           errorCode = "access_denied"
	ErrInvalidCredentials     errorCode = "invalid_credentials"
	ErrInvalidLogin           errorCode = "invalid_login"
	ErrInvalidRequest         errorCode = "invalid_request"
	ErrMethodNotAllowed       errorCode = "method_not_allowed"
	ErrNotFound               errorCode = "not_found"
	ErrServerError            errorCode = "server_error"
	ErrTemporarilyUnavailable errorCode = "temporarily_unavailable"
	ErrUnknown                errorCode = "unknown"
)

type errorCode string

func (c errorCode) Error() string {
	return c.String()
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

func NewError(code errorCode) *Error {
	e := &Error{ErrorCode: code}

	switch code {
	case ErrAccessDenied:
		e.statusCode = http.StatusUnauthorized
		e.ErrorDescription = "Your authentication request is denied."

	case ErrInvalidCredentials:
		e.statusCode = http.StatusUnauthorized
		e.ErrorDescription = "Authentication attempt failed."

	case ErrInvalidLogin:
		e.statusCode = http.StatusBadRequest
		e.ErrorDescription = "Invalid or malformed request."

	case ErrInvalidRequest:
		e.statusCode = http.StatusBadRequest
		e.ErrorDescription = "Invalid or malformed request."

	case ErrMethodNotAllowed:
		e.statusCode = http.StatusMethodNotAllowed
		e.ErrorDescription = "Invalid or malformed request."

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
				Msg("unknown auth error")

			code = ErrUnknown
		}

		e = NewError(code)
	}

	log.Debug().
		Int("status_code", e.StatusCode()).
		Str("error", e.Error()).
		Str("error_description", e.ErrorDescription).
		Msg("writing auth error")

	w.WriteHeader(e.statusCode)

	tmpl := errorTemplate()
	if err := tmpl.Execute(w, e); err != nil {
		log.Error().
			Err(err).
			Str("name", tmpl.Name()).
			Msg("failed to execute template")

		http.Error(w, e.ErrorDescription, e.statusCode)
	}
}

func writeJsonError(ctx context.Context, w http.ResponseWriter, err error) {
	log := zerolog.Ctx(ctx)

	var e *Error
	if !errors.As(err, &e) {
		var code errorCode
		if !errors.As(err, &code) {
			log.Warn().
				Err(err).
				Msg("unknown auth error")

			code = ErrUnknown
		}

		e = NewError(code)
	}

	log.Debug().
		Int("status_code", e.StatusCode()).
		Str("error", e.Error()).
		Str("error_description", e.ErrorDescription).
		Msg("writing json auth error")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.statusCode)

	b, err := json.Marshal(e)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal auth error")
		return
	}
	w.Write(b)
}
