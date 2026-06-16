package srv

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrAlreadyExists     errorCode = errorCode(codes.AlreadyExists)
	ErrConflict          errorCode = errorCode(codes.FailedPrecondition)
	ErrInternal          errorCode = errorCode(codes.Internal)
	ErrInvalidData       errorCode = errorCode(codes.InvalidArgument)
	ErrNotFound          errorCode = errorCode(codes.NotFound)
	ErrResourceExhausted errorCode = errorCode(codes.ResourceExhausted)
	ErrPermissionDenied  errorCode = errorCode(codes.PermissionDenied)
	ErrRevisionMismatch  errorCode = errorCode(codes.Aborted)
	ErrUnauthenticated   errorCode = errorCode(codes.Unauthenticated)
	ErrUnknown           errorCode = errorCode(codes.Unknown)
)

var (
	errInvalidAccessTokenPrefix = fmt.Errorf("invalid access token prefix")
	errInvalidAccessToken       = fmt.Errorf("invalid access token")
	errInvalidJwt               = fmt.Errorf("invalid jwt")
)

type errorCode codes.Code

func (c errorCode) Error() string {
	switch c {
	case ErrAlreadyExists:
		return "resource already exists"
	case ErrConflict:
		return "data conflict"
	case ErrInternal:
		return "internal error"
	case ErrInvalidData:
		return "invalid data"
	case ErrNotFound:
		return "resource not found"
	case ErrResourceExhausted:
		return "resource exhausted"
	case ErrRevisionMismatch:
		return "revision mismatch"
	case ErrPermissionDenied:
		return "forbidden"
	case ErrUnauthenticated:
		return "not authenticated"
	default:
		return "unknown error"
	}
}

type Error struct {
	code   errorCode
	err    error
	status *status.Status
}

func (e *Error) Code() errorCode {
	return e.code
}

func (e *Error) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	return e.code.Error()
}

func (e *Error) GRPCStatus() *status.Status {
	return e.status
}

func (e *Error) Unwrap() error {
	return e.err
}

func NewError(code errorCode, err error) error {
	return &Error{
		code:   code,
		err:    err,
		status: status.New(codes.Code(code), code.Error()),
	}
}

func NewErrorf(code errorCode, format string, a ...any) error {
	return &Error{
		code:   code,
		err:    fmt.Errorf(format, a...),
		status: status.New(codes.Code(code), fmt.Sprintf(format, a...)),
	}
}

func writeErrorHtml(ctx context.Context, w http.ResponseWriter, code int) {
	data := struct {
		Code    int
		Message string
	}{
		Code:    code,
		Message: http.StatusText(code),
	}

	w.WriteHeader(code)
	tmpl := errorTemplate()
	if err := tmpl.Execute(w, data); err != nil {
		log := zerolog.Ctx(ctx)
		log.Error().
			Err(err).
			Str("name", tmpl.Name()).
			Msg("failed to execute template")

		http.Error(w, data.Message, data.Code)
	}
}
