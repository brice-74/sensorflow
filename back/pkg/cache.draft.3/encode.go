package cache

import (
	"bytes"
	"encoding/gob"
)

func encodeGob[V any](v V) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}

func decodeGob[V any](data []byte, v *V) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
