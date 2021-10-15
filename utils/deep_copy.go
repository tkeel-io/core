package utils

import (
	"encoding/json"
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

	bytes, _ := json.Marshal(src)
	return json.Unmarshal(bytes, dst)
}

func Duplicate(v interface{}) interface{} {

	data := Data{}
	bytes, _ := json.Marshal(Data{Val: v})
	json.Unmarshal(bytes, &data)
	return data.Val
}
