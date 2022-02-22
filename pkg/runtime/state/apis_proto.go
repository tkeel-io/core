package state

import (
	"github.com/pkg/errors"
	"github.com/shamaton/msgpack/v2"
)

// ----------------- Request & Response.

type PatchData struct {
	Path     string      `json:"path"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// -------------------------------- Codec.

func GetPatchCodec() patchDataCodec { //nolint
	return patchDataCodec{}
}

type patchDataCodec struct {
}

func (c patchDataCodec) Encode(pds []PatchData) ([]byte, error) {
	bytes, err := msgpack.Marshal(pds)
	return bytes, errors.Wrap(err, "encode patch data")
}

func (c patchDataCodec) Decode(bytes []byte) ([]PatchData, error) {
	var pds []PatchData
	err := msgpack.Unmarshal(bytes, &pds)
	return pds, errors.Wrap(err, "decode patch data")
}
