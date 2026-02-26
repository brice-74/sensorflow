package parsestring

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// EvalCalc evaluates a simple arithmetic expression containing +, -, *, / operators and numeric literals.
// It respects operator precedence and supports both integer and floating-point types.
func EvalCalc[T constraints.Float | constraints.Integer](expr string) (T, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		var zero T
		return zero, errors.New("empty expression")
	}

	var nums []T
	var ops []rune

	start := 0
	for i, c := range expr {
		if c == '+' || c == '-' || c == '*' || c == '/' {
			num, err := Numeric[T](expr[start:i])
			if err != nil {
				return 0, err
			}
			nums = append(nums, num)

			if err := pushOp(&nums, &ops, c); err != nil {
				return 0, err
			}
			start = i + 1
		}
	}

	num, err := Numeric[T](expr[start:])
	if err != nil {
		return 0, err
	}
	nums = append(nums, num)

	for len(ops) > 0 {
		if err := reduceStacks(&nums, &ops); err != nil {
			return 0, err
		}
	}

	if len(nums) != 1 {
		return 0, errors.New("invalid expression")
	}

	return nums[0], nil
}

func pushOp[T constraints.Float | constraints.Integer](
	nums *[]T,
	ops *[]rune,
	op rune,
) error {
	for len(*ops) > 0 && hasPrecedence((*ops)[len(*ops)-1], op) {
		if err := reduceStacks(nums, ops); err != nil {
			return err
		}
	}
	*ops = append(*ops, op)
	return nil
}

func applyOp[T constraints.Float | constraints.Integer](a, b T, op rune) (T, error) {
	switch op {
	case '+':
		return a + b, nil
	case '-':
		return a - b, nil
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	default:
		return 0, errors.New("unknown operator")
	}
}

func reduceStacks[T constraints.Float | constraints.Integer](
	nums *[]T,
	ops *[]rune,
) error {
	if len(*nums) < 2 {
		return errors.New("invalid expression")
	}

	b := (*nums)[len(*nums)-1]
	a := (*nums)[len(*nums)-2]
	*nums = (*nums)[:len(*nums)-2]

	res, err := applyOp(a, b, (*ops)[len(*ops)-1])
	if err != nil {
		return err
	}

	*ops = (*ops)[:len(*ops)-1]
	*nums = append(*nums, res)
	return nil
}

func hasPrecedence(top, current rune) bool {
	if top == '*' || top == '/' {
		return true
	}
	return (top == '+' || top == '-') && (current == '+' || current == '-')
}

// Numeric parses a string into any numeric type T (int, uint, float).
// Supports float32, float64, int/int8/int16/int32/int64, uint/uint8/uint16/uint32/uint64.
// Returns an error if parsing fails or if the value overflows the target type.
func Numeric[T constraints.Float | constraints.Integer](s string) (T, error) {
	var zero T

	switch any(zero).(type) {
	// Floating point types
	case float32:
		f, err := strconv.ParseFloat(s, 32)
		return T(f), err
	case float64:
		f, err := strconv.ParseFloat(s, 64)
		return T(f), err

	// Signed integer types
	case int:
		i64, err := strconv.ParseInt(s, 10, 64)
		return T(i64), err
	case int8:
		i64, err := strconv.ParseInt(s, 10, 8)
		return T(i64), err
	case int16:
		i64, err := strconv.ParseInt(s, 10, 16)
		return T(i64), err
	case int32:
		i64, err := strconv.ParseInt(s, 10, 32)
		return T(i64), err
	case int64:
		i64, err := strconv.ParseInt(s, 10, 64)
		return T(i64), err

	// Unsigned integer types
	case uint:
		u64, err := strconv.ParseUint(s, 10, 64)
		return T(u64), err
	case uint8:
		u64, err := strconv.ParseUint(s, 10, 8)
		return T(u64), err
	case uint16:
		u64, err := strconv.ParseUint(s, 10, 16)
		return T(u64), err
	case uint32:
		u64, err := strconv.ParseUint(s, 10, 32)
		return T(u64), err
	case uint64:
		u64, err := strconv.ParseUint(s, 10, 64)
		return T(u64), err

	default:
		return zero, errors.New("unsupported numeric type")
	}
}
