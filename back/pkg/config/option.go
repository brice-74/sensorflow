package config

import (
	"fmt"
	"os"
)

type Option[T comparable] struct {
	flagName   string
	envName    string
	desc       string
	defaultVal T
	dst        *T
	flagPtr    *T
	parseEnv   func(string) (T, error)
	required   bool
}

// parse implements configOption for Option[T]
func (o *Option[T]) parse(mode Mode) error {
	switch mode {
	case FlagOnly, Both:
		// flagPtr always contains default or overridden by flag.Parse()
		*o.dst = *o.flagPtr
	case EnvOnly:
		if raw := os.Getenv(o.envName); raw != "" {
			parsed, err := o.parseEnv(raw)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", o.envName, err)
			}
			*o.dst = parsed
		} else {
			*o.dst = o.defaultVal
		}
	}

	if o.required && isZero(*o.dst) {
		return fmt.Errorf("missing required config %s (env=%s)", o.flagName, o.envName)
	}

	return nil
}

func (o *Option[T]) Required() *Option[T] {
	o.required = true
	return o
}

func isZero[T comparable](val T) bool {
	var zero T
	return val == zero
}
