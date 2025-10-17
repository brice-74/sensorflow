package gobutil

import (
	"bytes"
	"encoding/gob"
)

func Encode[V any](value *V) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode[V any](data []byte) (*V, error) {
	var v V
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func DecodeManyAnyStr[V any](values []any) ([]*V, error) {
	result := make([]*V, 0, len(values))
	for _, val := range values {
		s, ok := val.(string)
		if !ok || s == "" {
			continue
		}

		v, err := Decode[V]([]byte(s))
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
