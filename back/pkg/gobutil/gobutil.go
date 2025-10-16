package gobutil

import (
	"bytes"
	"encoding/gob"
)

func Decode[V any](data []byte) (*V, error) {
	var v V
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func Encode[V any](value *V) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
