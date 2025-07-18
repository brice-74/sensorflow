package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

type Mode uint8

const (
	EnvOnly Mode = iota
	FlagOnly
	Both // flags override env
)

type configOption interface {
	parse(mode Mode) error
}

type Loader struct {
	mode    Mode
	options []configOption
}

func NewLoader(mode Mode) *Loader {
	return &Loader{mode: mode}
}

func (l *Loader) Parse() error {
	if l.mode != EnvOnly {
		flag.Parse()
	}

	var errs []error
	for _, opt := range l.options {
		if err := opt.parse(l.mode); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		var sb strings.Builder
		sb.WriteString("config errors:\n")
		for _, err := range errs {
			sb.WriteString(" - ")
			sb.WriteString(err.Error())
			sb.WriteByte('\n')
		}
		return errors.New(sb.String())
	}
	return nil
}

// addOption registers a generic Option[T]
func addOption[T comparable](
	l *Loader,
	dst *T,
	flagName, envName, desc string,
	defaultVal T,
	// registerFlag should be flag.String, flag.Float64, etc.
	registerFlag func(name string, def T, desc string) *T,
	// parseEnv converts a string to T
	parseEnv func(string) (T, error),
) *Option[T] {
	if dst == nil {
		panic("config: destination pointer is nil")
	}

	var flagDefault T
	if l.mode == EnvOnly || l.mode == Both {
		if raw := os.Getenv(envName); raw != "" {
			if parsed, err := parseEnv(raw); err == nil {
				flagDefault = parsed
			} else {
				flagDefault = defaultVal
			}
		} else {
			flagDefault = defaultVal
		}
	} else {
		flagDefault = defaultVal
	}

	flagPtr := registerFlag(flagName, flagDefault, desc)

	opt := &Option[T]{
		flagName:   flagName,
		envName:    envName,
		desc:       desc,
		defaultVal: defaultVal,
		flagPtr:    flagPtr,
		parseEnv:   parseEnv,
		dst:        dst,
	}
	l.options = append(l.options, opt)
	return opt
}

func (l *Loader) String(dst *string, flagName, envName, defaultVal, desc string) *Option[string] {
	return addOption(l, dst, flagName, envName, desc, defaultVal,
		flag.String,
		func(s string) (string, error) { return s, nil },
	)
}

func (l *Loader) Bool(dst *bool, flagName, envName string, defaultVal bool, desc string) *Option[bool] {
	return addOption(l, dst, flagName, envName, desc, defaultVal,
		flag.Bool,
		func(s string) (bool, error) { return strconv.ParseBool(s) },
	)
}

func (l *Loader) Int(dst *int, flagName, envName string, defaultVal int, desc string) *Option[int] {
	return addOption(l, dst, flagName, envName, desc, defaultVal,
		flag.Int,
		func(s string) (int, error) {
			v, err := strconv.ParseInt(s, 10, 0)
			return int(v), err
		},
	)
}

func (l *Loader) Float64(dst *float64, flagName, envName string, defaultVal float64, desc string) *Option[float64] {
	return addOption(l, dst, flagName, envName, desc, defaultVal,
		flag.Float64,
		func(s string) (float64, error) { return strconv.ParseFloat(s, 64) },
	)
}

func (l *Loader) Uint(dst *uint, flagName, envName string, defaultVal uint, desc string) *Option[uint] {
	return addOption(l, dst, flagName, envName, desc, defaultVal,
		flag.Uint,
		func(s string) (uint, error) {
			u64, err := strconv.ParseUint(s, 10, 64)
			return uint(u64), err
		},
	)
}

func (l *Loader) Duration(dst *time.Duration, flagName, envName string, defaultVal time.Duration, desc string) *Option[time.Duration] {
	return addOption(l, dst, flagName, envName, desc, defaultVal,
		flag.Duration,
		time.ParseDuration,
	)
}
