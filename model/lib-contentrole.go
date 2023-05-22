package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type ContentRole int

const (
	ContentRoleNone       = ContentRole(0)
	ContentRoleSource     = ContentRole(1)
	ContentRoleTranscript = ContentRole(2)
	ContentRoleMemory     = ContentRole(3)
	ContentRoleSummary    = ContentRole(4)
)

var _contentRoleStrings = []string{
	"",
	"source",
	"transcript",
	"memory",
	"summary",
}

func (v ContentRole) String() string {
	return _contentRoleStrings[v]
}

func ParseContentRole(s string) (ContentRole, error) {
	if i := slices.Index(_contentRoleStrings, s); i >= 0 {
		return ContentRole(i), nil
	} else {
		return ContentRoleNone, fmt.Errorf("invalid ContentRole %q", s)
	}
}

func (v ContentRole) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *ContentRole) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseContentRole(string(b))
	return err
}
func (v ContentRole) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *ContentRole) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = ContentRole(n)
	return err
}
