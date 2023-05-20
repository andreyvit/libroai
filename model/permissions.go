package m

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Permission int

const (
	PermissionNone = Permission(iota)

	PermissionAccessSuperadminArea
	PermissionManageSuperadmins

	PermissionAccessAdminArea
	PermissionManageAccount

	PermissionAccessChat
)

var _permissionStrings = []string{
	"none",

	"access-superadmin-area",
	"manage-superadmins",

	"access-admin-area",
	"manage-admins",

	"access-chat",
}

func (v Permission) String() string {
	return _permissionStrings[v]
}

func ParsePermission(s string) (Permission, error) {
	if i := slices.Index(_permissionStrings, s); i >= 0 {
		return Permission(i), nil
	} else {
		return PermissionNone, fmt.Errorf("invalid Permission %q", s)
	}
}

func (v Permission) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Permission) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParsePermission(string(b))
	return err
}
