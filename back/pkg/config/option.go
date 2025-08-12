package config

import (
	"flag"
	"fmt"
	"os"
)

type Option[T comparable] struct {
	flagName     string
	envName      string
	desc         string
	defaultVal   T
	dst          *T
	flagPtr      *string
	fromString   func(string) (T, error)
	required     bool
	validationFn func(T) error
}

// parse implements configOption for Option[T]
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

	isZero := isZero(*o.dst)
	if o.required && isZero {
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

func isZero[T comparable](val T) bool {
	var zero T
	return val == zero
}

// AddOption registers a generic Option[T]
func AddOption[T comparable](
	l *Loader,
	dst *T,
	flagName, envName, desc string,
	defaultVal T,
	// parseEnv converts a string to T
	fromString func(string) (T, error),
) *Option[T] {
	if dst == nil {
		panic("config: destination pointer is nil")
	}

	opt := &Option[T]{
		flagName:   flagName,
		envName:    envName,
		desc:       desc,
		defaultVal: defaultVal,
		dst:        dst,
		fromString: fromString,
	}

	if l.mode != EnvOnly {
		opt.flagPtr = flag.String(flagName, "", desc)
	}

	l.options = append(l.options, opt)
	return opt
}
