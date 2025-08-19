package parsestring

import (
	"errors"
	"strings"

	"golang.org/x/exp/constraints"
)

type byteMultiplier int64

const (
	B  byteMultiplier = 1
	KB                = 1024 * B
	MB                = 1024 * KB
	GB                = 1024 * MB
	TB                = 1024 * GB
)

// ByteSize converts strings such as "1.5MB", "1024B", and "2GiB" into bytes.
// Only supports standard SI/IEC conventions.
func ByteSize[T constraints.Float | constraints.Integer](s string) (T, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		var zero T
		return zero, errors.New("empty size string")
	}

	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && c != '.' {
			break
		}
	}
	if i == 0 {
		var zero T
		return zero, errors.New("no numeric part found")
	}

	numStr := s[:i]
	suffix := strings.ToUpper(strings.TrimSpace(s[i:]))

	num, err := Numeric[T](numStr)
	if err != nil {
		var zero T
		return zero, err
	}

	var multiplier byteMultiplier
	switch suffix {
	case "B", "":
		multiplier = B
	case "KB", "KIB":
		multiplier = KB
	case "MB", "MIB":
		multiplier = MB
	case "GB", "GIB":
		multiplier = GB
	case "TB", "TIB":
		multiplier = TB
	default:
		var zero T
		return zero, errors.New("invalid size suffix: " + suffix)
	}

	return num * T(multiplier), nil
}
