package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type EmbeddingType uint8

const (
	EmbeddingTypeNone  = EmbeddingType(0)
	EmbeddingTypeAda02 = EmbeddingType(1)
)

var _embeddingTypeStrings = []string{
	"",
	"ada02",
}

func (v EmbeddingType) String() string {
	return _embeddingTypeStrings[v]
}

func ParseEmbeddingType(s string) (EmbeddingType, error) {
	if i := slices.Index(_embeddingTypeStrings, s); i >= 0 {
		return EmbeddingType(i), nil
	} else {
		return EmbeddingTypeNone, fmt.Errorf("invalid EmbeddingType %q", s)
	}
}

func (v EmbeddingType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *EmbeddingType) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseEmbeddingType(string(b))
	return err
}
func (v EmbeddingType) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *EmbeddingType) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = EmbeddingType(n)
	return err
}
