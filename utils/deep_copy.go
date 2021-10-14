package utils

import (
	"bytes"
	"encoding/gob"
)

func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(src)
	if err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func Duplicate(v interface{}) (interface{}, error) {
	copyV := new(interface{})
	return *copyV, DeepCopy(copyV, &v)
}
