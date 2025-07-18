package ulid

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"
)

var (
	entropy    = ulid.Monotonic(rand.Reader, 0)
	ErrInvalid = errors.New("invalid ULID")
)

type ULID struct{ ulid.ULID }

func New() ULID {
	return ULID{ulid.Make()}
}

func NewOrdered() ULID {
	return ULID{ulid.MustNew(ulid.Now(), entropy)}
}

func (id ULID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *ULID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("%w: expected string, got %s", ErrInvalid, string(data))
	}
	parsed, err := ulid.Parse(s)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalid, s)
	}
	id.ULID = parsed
	return nil
}
