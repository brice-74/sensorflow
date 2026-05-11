package config

import (
	"cmp"
	"fmt"
	"strconv"
)

func IsGreaterThan[T cmp.Ordered](min T) func(T) error {
	return func(value T) error {
		if value <= min {
			return fmt.Errorf("must be > %v", min)
		}
		return nil
	}
}

func IsGreaterEqualThan[T cmp.Ordered](min T) func(T) error {
	return func(value T) error {
		if value < min {
			return fmt.Errorf("must be >= %v", min)
		}
		return nil
	}
}

func IsLessThan[T cmp.Ordered](max T) func(T) error {
	return func(value T) error {
		if value >= max {
			return fmt.Errorf("must be < %v", max)
		}
		return nil
	}
}

func IsLessEqualThan[T cmp.Ordered](max T) func(T) error {
	return func(value T) error {
		if value > max {
			return fmt.Errorf("must be <= %v", max)
		}
		return nil
	}
}

func IsBetween[T cmp.Ordered](min, max T) func(T) error {
	return func(value T) error {
		if value < min || value > max {
			return fmt.Errorf("must be between %v and %v", min, max)
		}
		return nil
	}
}

func IsNonEmptyString(value string) error {
	if value == "" {
		return fmt.Errorf("must not be empty")
	}
	return nil
}

func IsPortString(value string) error {
	portNum, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	return IsBetween(1, 65535)(portNum)
}
