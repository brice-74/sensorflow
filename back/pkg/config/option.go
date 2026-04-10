package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Option[T any] struct {
	flagName     string
	envName      string
	desc         string
	defaultVal   T
	dst          *T
	flagPtr      *string
	fromString   func(string) (T, error)
	required     bool
	validationFn func(T) error
	isZeroFn     func(T) bool
}

var _ option = (*Option[any])(nil)

func (o *Option[T]) parse(mode Mode) error {
	var raw string
	switch mode {
	case EnvOnly:
		raw = os.Getenv(o.envName)
	case FlagOnly:
		raw = *o.flagPtr
	case Both:
		if *o.flagPtr != "" && *o.flagPtr != fmt.Sprint(o.defaultVal) {
			raw = *o.flagPtr
		} else {
			raw = os.Getenv(o.envName)
		}
	}

	if raw == "" {
		*o.dst = o.defaultVal
	} else {
		val, err := o.fromString(raw)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", o.envName, err)
		}
		*o.dst = val
	}

	if o.required && o.isZeroFn(*o.dst) {
		return fmt.Errorf("missing required config %s (env=%s)", o.flagName, o.envName)
	}

	if o.validationFn != nil {
		if err := o.validationFn(*o.dst); err != nil {
			return fmt.Errorf(
				"validation config %s (env=%s): value %v, error: %s",
				o.flagName, o.envName, *o.dst, err,
			)
		}
	}

	return nil
}

func (o *Option[T]) Required() *Option[T] {
	o.required = true
	return o
}

func (o *Option[T]) Validate(fn func(T) error) *Option[T] {
	o.validationFn = fn
	return o
}

func isZero[T any](val T) bool {
	return reflect.ValueOf(val).IsZero()
}

// AddOption registers a generic Option[T]
func AddOption[T any](
	l *Loader,
	dst *T,
	flagName, envName string,
	defaultVal T,
	desc string,
	fromString func(string) (T, error),
) *Option[T] {
	if dst == nil {
		panic("config: destination pointer is nil")
	}

	flagName = l.prefixed(flagName, 0)
	envName = l.prefixed(envName, 1)
	opt := &Option[T]{
		flagName:   flagName,
		envName:    envName,
		desc:       desc,
		defaultVal: defaultVal,
		dst:        dst,
		fromString: fromString,
		isZeroFn:   isZero[T],
	}

	if l.mode != EnvOnly {
		opt.flagPtr = flag.String(flagName, "", desc)
	}

	if l.options == nil {
		panic("config: loader options storage is nil")
	}
	*l.options = append(*l.options, opt)
	return opt
}

func AddSliceOption[T any](
	l *Loader,
	dst *[]T,
	flagName, envName string,
	defaultVal []T,
	desc string,
	separator string,
	fromStringIter func(string) (T, error),
) *Option[[]T] {
	opt := AddOption(l, dst, flagName, envName, defaultVal, desc, func(s string) ([]T, error) {
		parts := strings.Split(s, separator)
		var result []T
		for _, part := range parts {
			val, err := fromStringIter(part)
			if err != nil {
				return nil, err
			}
			result = append(result, val)
		}
		return result, nil
	})

	opt.isZeroFn = func(val []T) bool {
		return len(val) == 0
	}

	return opt
}
