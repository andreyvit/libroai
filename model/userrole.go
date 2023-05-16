package m

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type UserSystemRole int

const (
	UserSystemRoleNone       = UserSystemRole(0)
	UserSystemRoleRegular    = UserSystemRole(1)
	UserSystemRoleSuperadmin = UserSystemRole(2)
	UserSystemRoleSuperQA    = UserSystemRole(3)
)

func (role UserSystemRole) IsSuper() bool {
	return role >= UserSystemRoleSuperadmin
}

var _userSystemRoleStrings = []string{
	"none",
	"regular",
	"superadmin",
	"superqa",
}

type UserAccountRole int

const (
	UserAccountRoleNone      = UserAccountRole(0)
	UserAccountRoleConsumer  = UserAccountRole(1)
	UserAccountRoleOwner     = UserAccountRole(2)
	UserAccountRoleAdmin     = UserAccountRole(3)
	UserAccountRoleAssistant = UserAccountRole(4)
)

func (role UserAccountRole) HasBackofficeAccess() bool {
	return role >= UserAccountRoleOwner
}

var _userAccountRoleStrings = []string{
	"none",
	"consumer",
	"owner",
	"admin",
	"assistant",
}

func (v UserSystemRole) String() string {
	return _userSystemRoleStrings[v]
}
func ParseUserSystemRole(s string) (UserSystemRole, error) {
	if i := slices.Index(_userSystemRoleStrings, s); i >= 0 {
		return UserSystemRole(i), nil
	} else {
		return UserSystemRoleNone, fmt.Errorf("invalid UserSystemRole %q", s)
	}
}
func (v UserSystemRole) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *UserSystemRole) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseUserSystemRole(string(b))
	return err
}
func (v UserSystemRole) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint8(uint8(v))
}
func (v *UserSystemRole) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint8()
	*v = UserSystemRole(n)
	return err
}

func (v UserAccountRole) String() string {
	return _userAccountRoleStrings[v]
}
func ParseUserAccountRole(s string) (UserAccountRole, error) {
	if i := slices.Index(_userAccountRoleStrings, s); i >= 0 {
		return UserAccountRole(i), nil
	} else {
		return UserAccountRoleNone, fmt.Errorf("invalid UserAccountRole %q", s)
	}
}
func (v UserAccountRole) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *UserAccountRole) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseUserAccountRole(string(b))
	return err
}
func (v UserAccountRole) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint8(uint8(v))
}
func (v *UserAccountRole) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint8()
	*v = UserAccountRole(n)
	return err
}
