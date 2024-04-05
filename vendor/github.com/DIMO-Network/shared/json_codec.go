package shared

import "encoding/json"

// TODO(elffjs): Move this to DIMO-Network/shared.
type JSONCodec[A any] struct{}

func (c *JSONCodec[A]) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *JSONCodec[A]) Decode(data []byte) (interface{}, error) {
	value := new(A)
	return value, json.Unmarshal(data, value)
}
