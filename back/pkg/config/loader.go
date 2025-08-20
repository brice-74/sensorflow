package config

import (
	"errors"
	"flag"
	"strconv"
	"time"
)

type Configurable interface {
	Define(loader *Loader)
}

type Mode uint8

const (
	EnvOnly Mode = iota
	FlagOnly
	Both // flags override env
)

type option interface {
	parse(mode Mode) error
}

type Loader struct {
	mode     Mode
	options  []option
	prefixes [2]string // 0 = flag prefix, 1 = env prefix
}

func NewLoader(mode Mode, configurators ...Configurable) *Loader {
	loader := &Loader{mode: mode}
	for _, c := range configurators {
		c.Define(loader)
	}
	return loader
}

func (l *Loader) Prefixes() [2]string {
	return l.prefixes
}

func (l *Loader) WithPrefixes(flagPrefix, envPrefix string) *Loader {
	clone := *l
	clone.prefixes = [2]string{flagPrefix, envPrefix}
	return &clone
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
		return errors.Join(errs...)
	}
	return nil
}

func (l *Loader) String(dst *string, flagName, envName, defaultVal, desc string, prefix ...string) *Option[string] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc,
		func(s string) (string, error) { return s, nil },
	)
}

func (l *Loader) Bool(dst *bool, flagName, envName string, defaultVal bool, desc string, prefix ...string) *Option[bool] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc,
		func(s string) (bool, error) { return strconv.ParseBool(s) },
	)
}

func (l *Loader) Int(dst *int, flagName, envName string, defaultVal int, desc string, prefix ...string) *Option[int] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc,
		func(s string) (int, error) {
			v, err := strconv.ParseInt(s, 10, 0)
			return int(v), err
		},
	)
}

func (l *Loader) Float64(dst *float64, flagName, envName string, defaultVal float64, desc string, prefix ...string) *Option[float64] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc,
		func(s string) (float64, error) { return strconv.ParseFloat(s, 64) },
	)
}

func (l *Loader) Uint(dst *uint, flagName, envName string, defaultVal uint, desc string, prefix ...string) *Option[uint] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc,
		func(s string) (uint, error) {
			u64, err := strconv.ParseUint(s, 10, 64)
			return uint(u64), err
		},
	)
}

func (l *Loader) Uint16(dst *uint16, flagName, envName string, defaultVal uint16, desc string, prefix ...string) *Option[uint16] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc,
		func(s string) (uint16, error) {
			u64, err := strconv.ParseUint(s, 10, 16)
			return uint16(u64), err
		},
	)
}

func (l *Loader) Duration(dst *time.Duration, flagName, envName string, defaultVal time.Duration, desc string, prefix ...string) *Option[time.Duration] {
	return AddOption(l, dst, flagName, envName, defaultVal, desc, time.ParseDuration)
}
