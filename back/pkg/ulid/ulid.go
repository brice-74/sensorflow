package ulid

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"
)

var (
	Zero       ULID
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

func ToStrings(ids []ULID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

// Helpers to avoid importing the stdlib "github.com/oklog/ulid/v2".

func Parse(ulidstr string) (ULID, error) {
	id, err := ulid.Parse(ulidstr)
	return ULID{id}, err
}

func ParseStrict(ulidstr string) (ULID, error) {
	id, err := ulid.ParseStrict(ulidstr)
	return ULID{id}, err
}
