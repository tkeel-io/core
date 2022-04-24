package dao

type Codec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(raw []byte, v interface{}) error
}

type KVCodec interface {
	Key() Codec
	Value() Codec
}
