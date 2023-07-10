package m

import (
	"errors"
	"time"

	"github.com/andreyvit/mvp/flake"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

type (
	UserID = flake.ID

	User struct {
		ID          UserID            `msgpack:"-"`
		Memberships []*UserMembership `msgpack:"a"`
		Role        UserSystemRole    `msgpack:"r"`
		Email       string            `msgpack:"e"`
		EmailNorm   string            `msgpack:"e!"`
		Name        string            `msgpack:"n"`
		LoginMsg    string            `msgpack:"msg,omitempty"`
	}

	UserMembershipID = flake.ID

	UserMembership struct {
		CreationTime time.Time       `msgpack:"@"`
		AccountID    AccountID       `msgpack:"a"`
		Role         UserAccountRole `msgpack:"r"`
		Status       UserStatus      `msgpack:"t"`
		Source       UserSource      `msgpack:"s,omitempty"`
		Comment      UserSource      `msgpack:"c,omitempty"`
	}
)

func (obj *User) FlakeID() flake.ID {
	return obj.ID
}
func (User) ObjectType() mvpm.Type {
	return mvpm.TypeUser
}

func (u *User) Membership(accountID AccountID) *UserMembership {
	for _, m := range u.Memberships {
		if m.AccountID == accountID {
			return m
		}
	}
	return nil
}

func (u *User) MembershipRole(accountID AccountID) UserAccountRole {
	if accountID == 0 {
		return UserAccountRoleNone
	}
	if m := u.Membership(accountID); m != nil {
		return m.Role
	} else {
		return UserAccountRoleNone
	}
}

// func (u *User) Check(perm Permission, accountID AccountID, obj mvpm.Object) error {
// 	return CheckAccess(u, perm, accountID, obj)
// }

type UserSource int

const (
	UserSourceDefault   = UserSource(0)
	UserSourceWhitelist = UserSource(1)
)

var (
	ErrForbiddenNotSuperadmin = errors.New("Only superadmins can access this area.")
	ErrForbiddenWrongAccount  = errors.New("You do not have access to this account.")
	ErrForbiddenNotStaff      = errors.New("Only staff can access this area.")
	ErrForbiddenOther         = errors.New("Forbidden.")
)

func CanAccess(u *User, perm Permission, accountID AccountID, obj mvpm.Object) bool {
	return CheckAccess(u, perm, accountID, obj) == nil
}

func CheckAccess(u *User, perm Permission, accountID AccountID, obj mvpm.Object) error {
	ar := u.MembershipRole(accountID)

	switch perm {
	case PermissionSwitchToAccount:
		if accountID == 0 {
			panic("zero account ID")
		}
		if u.Role.IsSuper() {
			return nil
		}
		if ar != UserAccountRoleNone {
			return nil
		}
		return ErrForbiddenWrongAccount
	case PermissionAccessSuperadminArea:
		if u.Role.IsSuper() {
			return nil
		}
		return ErrForbiddenNotSuperadmin
	case PermissionManageSuperadmins:
		if u.Role == UserSystemRoleSuperadmin {
			return nil
		}
		return ErrForbiddenNotSuperadmin
	case PermissionAccessAdminArea:
		if accountID == 0 {
			panic("zero account ID")
		}
		if u.Role == UserSystemRoleSuperadmin {
			return nil
		}
		switch ar {
		case UserAccountRoleOwner, UserAccountRoleAdmin, UserAccountRoleAssistant:
			return nil
		case UserAccountRoleNone:
			return ErrForbiddenWrongAccount
		case UserAccountRoleConsumer:
			return ErrForbiddenNotStaff
		}
	case PermissionManageAccount:
		if accountID == 0 {
			panic("zero account ID")
		}
		if u.Role == UserSystemRoleSuperadmin {
			return nil
		}
		switch ar {
		case UserAccountRoleOwner, UserAccountRoleAdmin:
			return nil
		case UserAccountRoleNone:
			return ErrForbiddenWrongAccount
		case UserAccountRoleConsumer:
			return ErrForbiddenNotStaff
		}
	}
	return ErrForbiddenOther
}
