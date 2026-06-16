package db

import "errors"

const (
	ErrUnknown          errorCode = 0
	ErrNotFound         errorCode = 1
	ErrRevisionMismatch errorCode = 2
	ErrInvalidData      errorCode = 3
	ErrConflict         errorCode = 4
)

type errorCode uint32

func (c errorCode) Error() string {
	switch c {
	case ErrConflict:
		return "data conflict"
	case ErrInvalidData:
		return "invalid data"
	case ErrNotFound:
		return "resource not found"
	case ErrRevisionMismatch:
		return "revision mismatch"
	default:
		return "unknown error"
	}
}

type wrappedError struct {
	visible error
	wrapped error
}

func (e *wrappedError) Error() string {
	return e.visible.Error()
}

func (e *wrappedError) Is(target error) bool {
	return errors.Is(e.visible, target)
}

func (e *wrappedError) Unwrap() error {
	return e.wrapped
}

func wrapError(wrapped, visible error) error {
	return &wrappedError{
		visible: visible,
		wrapped: wrapped,
	}
}
