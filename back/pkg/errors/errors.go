package errors

import "fmt"

type Code = uint8

const (
	ErrInternal Code = iota
	ErrNotFound
	ErrAlreadyExists
	ErrInvalidReference
	ErrInvalidInput
	ErrUnexpectedRows
	ErrTimeout
	ErrCanceled
)

func CodeToString(c Code) string {
	switch c {
	case ErrNotFound:
		return "NotFound"
	case ErrAlreadyExists:
		return "AlreadyExists"
	case ErrInvalidReference:
		return "InvalidReference"
	case ErrInvalidInput:
		return "InvalidInput"
	case ErrUnexpectedRows:
		return "UnexpectedRows"
	case ErrTimeout:
		return "Timeout"
	case ErrCanceled:
		return "Canceled"
	default:
		return "Internal"
	}
}

type Error struct {
	Code    Code
	Err     error
	Details map[string]any
}

func (e *Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("code: %s, error: %v, details: %v", CodeToString(e.Code), e.Err, e.Details)
	}
	return fmt.Sprintf("code: %s, error: %v", CodeToString(e.Code), e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewError(code Code, err error, details map[string]any) *Error {
	return &Error{
		Code:    code,
		Err:     err,
		Details: details,
	}
}

func NewWrappedErr(code Code, err error, details map[string]any) *Error {
	return &Error{
		Code:    code,
		Err:     WrapErr(err),
		Details: details,
	}
}
