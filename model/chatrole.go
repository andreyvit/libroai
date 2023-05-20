package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type MessageRole int

const (
	MessageRoleNone   = MessageRole(0)
	MessageRoleUser   = MessageRole(1)
	MessageRoleBot    = MessageRole(2)
	MessageRoleSystem = MessageRole(3)
)

func (v MessageRole) IsUser() bool {
	return v == MessageRoleUser
}
func (v MessageRole) IsBot() bool {
	return v == MessageRoleBot
}
func (v MessageRole) IsSystem() bool {
	return v == MessageRoleSystem
}

var _messageRoleStrings = []string{
	"none",
	"user",
	"bot",
	"system",
}

func (v MessageRole) String() string {
	return _messageRoleStrings[v]
}

func ParseMessageRole(s string) (MessageRole, error) {
	if i := slices.Index(_messageRoleStrings, s); i >= 0 {
		return MessageRole(i), nil
	} else {
		return MessageRoleNone, fmt.Errorf("invalid MessageRole %q", s)
	}
}

func (v MessageRole) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *MessageRole) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseMessageRole(string(b))
	return err
}
func (v MessageRole) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *MessageRole) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = MessageRole(n)
	return err
}
