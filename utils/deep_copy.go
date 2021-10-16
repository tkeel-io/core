package utils

import (
	"encoding/json"
	"fmt"
)

type Data struct {
	Val interface{} `json:"val"`
}

// func DeepCopy(dst, src interface{}) error {
// 	var buf bytes.Buffer
// 	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
// 		return errors.Unwrap(err)
// 	}
// 	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst); err != nil {
// 		return errors.Unwrap(err)
// 	}
// 	return nil
// }

func DeepCopy(dst, src interface{}) error {
	var (
		err   error
		bytes []byte
	)

	if bytes, err = json.Marshal(src); nil != err {
		return fmt.Errorf("json marshal error, %w", err)
	}
	if err = json.Unmarshal(bytes, dst); nil != err {
		return fmt.Errorf("json marshal error, %w", err)
	}

	return nil
}

func Duplicate(v interface{}) interface{} {
	data := Data{}
	bytes, _ := json.Marshal(Data{Val: v})
	json.Unmarshal(bytes, &data)
	return data.Val
}
