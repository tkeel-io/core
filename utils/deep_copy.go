package utils

import (
	"bytes"
	"encoding/gob"
	"errors"
)

func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return errors.Unwrap(err)
	}
	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst); err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func Duplicate(v interface{}) (interface{}, error) {
	copyV := new(interface{})
	return *copyV, DeepCopy(copyV, &v)
}
