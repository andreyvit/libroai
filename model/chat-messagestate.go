package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type MessageState int

const (
	MessageStateFinished = MessageState(0)
	MessageStatePending  = MessageState(1)
	MessageStateFailed   = MessageState(2)
)

var _messageStateStrings = []string{
	"finished",
	"pending",
	"failed",
}

func (v MessageState) IsPending() bool {
	return v == MessageStatePending
}
func (v MessageState) IsFailed() bool {
	return v == MessageStateFailed
}

func (v MessageState) String() string {
	return _messageStateStrings[v]
}
func ParseMessageState(s string) (MessageState, error) {
	if i := slices.Index(_messageStateStrings, s); i >= 0 {
		return MessageState(i), nil
	} else {
		return MessageStateFinished, fmt.Errorf("invalid MessageState %q", s)
	}
}
func (v MessageState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *MessageState) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseMessageState(string(b))
	return err
}
func (v MessageState) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *MessageState) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = MessageState(n)
	return err
}
