package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type UserStatus int

const (
	UserStatusUnknown      = UserStatus(0)
	UserStatusActive       = UserStatus(1)
	UserStatusInactive     = UserStatus(2)
	UserStatusBanned       = UserStatus(3)
	UserStatusInvited      = UserStatus(4)
	UserStatusSelfRejected = UserStatus(5)
)

var _userStatusStrings = []string{
	"unknown",
	"active",
	"inactive",
	"banned",
	"invited",
	"selfrejected",
}

func (v UserStatus) IsKnown() bool {
	return v != UserStatusUnknown
}

func (v UserStatus) Invitable() bool {
	return v == UserStatusUnknown || v == UserStatusInactive || v == UserStatusBanned
}

func (v UserStatus) ActiveOrInvited() bool {
	return v == UserStatusActive || v == UserStatusInvited
}

func (v UserStatus) String() string {
	return _userStatusStrings[v]
}
func ParseUserStatus(s string) (UserStatus, error) {
	if i := slices.Index(_userStatusStrings, s); i >= 0 {
		return UserStatus(i), nil
	} else {
		return UserStatusUnknown, fmt.Errorf("invalid UserStatus %q", s)
	}
}
func (v UserStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *UserStatus) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseUserStatus(string(b))
	return err
}
func (v UserStatus) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint8(uint8(v))
}
func (v *UserStatus) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint8()
	*v = UserStatus(n)
	return err
}
