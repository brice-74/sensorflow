package errors

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// used to avoid importing the stdlib "errors".
var (
	Is     = errors.Is
	As     = errors.As
	Unwrap = errors.Unwrap
	Join   = errors.Join
)

var pcNameCache sync.Map

func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "(unknown caller)"
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func wrap(err error, msg string, skip int) error {
	if err == nil {
		return nil
	}
	fn := GetCaller(skip)
	if msg == "" {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return fmt.Errorf("%s: %s: %w", fn, msg, err)
}

func Wrap(err error, msg string) error {
	return wrap(err, msg, 3)
}

func Wrapf(err error, format string, args ...any) error {
	return wrap(err, fmt.Sprintf(format, args...), 3)
}

func WrapErr(err error) error {
	return wrap(err, "", 3)
}

func WrapMsg(msg string) error {
	return wrap(errors.New(msg), "", 3)
}

func WrapMsgf(format string, args ...any) error {
	return wrap(fmt.Errorf(format, args...), "", 3)
}

func JoinWrap(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}

	fn := GetCaller(2)
	joined := errors.Join(errs...)

	return fmt.Errorf("%s: %w", fn, joined)
}
