package errors

import (
	"fmt"
)

var (
	ErrInternal         error = &Error{Code: CodeInternal}
	ErrNotFound         error = &Error{Code: CodeNotFound}
	ErrAlreadyExists    error = &Error{Code: CodeAlreadyExists}
	ErrInvalidReference error = &Error{Code: CodeInvalidReference}
	ErrInvalidInput     error = &Error{Code: CodeInvalidInput}
	ErrUnexpectedRows   error = &Error{Code: CodeUnexpectedRows}
	ErrTimeout          error = &Error{Code: CodeTimeout}
	ErrCanceled         error = &Error{Code: CodeCanceled}
)

type Code uint8

const (
	CodeInternal Code = iota
	CodeNotFound
	CodeAlreadyExists
	CodeInvalidReference
	CodeInvalidInput
	CodeUnexpectedRows
	CodeTimeout
	CodeCanceled
)

func (c Code) String() string {
	switch c {
	case CodeNotFound:
		return "NotFound"
	case CodeAlreadyExists:
		return "AlreadyExists"
	case CodeInvalidReference:
		return "InvalidReference"
	case CodeInvalidInput:
		return "InvalidInput"
	case CodeUnexpectedRows:
		return "UnexpectedRows"
	case CodeTimeout:
		return "Timeout"
	case CodeCanceled:
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
		return fmt.Sprintf("code: %s, error: %v, details: %v", e.Code, e.Err, e.Details)
	}
	return fmt.Sprintf("code: %s, error: %v", e.Code, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
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
