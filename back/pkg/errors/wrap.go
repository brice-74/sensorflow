package errors

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
)

var pcNameCache sync.Map

func getCallerFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "(unknown caller)"
	}

	if name, ok := pcNameCache.Load(pc); ok {
		return name.(string)
	}

	name := runtime.FuncForPC(pc).Name()
	actual, _ := pcNameCache.LoadOrStore(pc, name)
	return actual.(string)
}

func wrap(err error, msg string, skip int) error {
	if err == nil {
		return nil
	}
	fn := getCallerFuncName(skip)
	if msg == "" {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return fmt.Errorf("%s: %s: %w", fn, msg, err)
}

func Wrap(err error, msg string) error {
	return wrap(err, msg, 2)
}

func Wrapf(err error, format string, args ...any) error {
	return wrap(err, fmt.Sprintf(format, args...), 2)
}

func WrapErr(err error) error {
	return wrap(err, "", 2)
}

func New(msg string) error {
	return wrap(errors.New(msg), "", 2)
}

func Newf(format string, args ...any) error {
	return wrap(fmt.Errorf(format, args...), "", 2)
}
