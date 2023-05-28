package m

import (
	"time"

	"github.com/andreyvit/mvp/flake"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

const (
	TypeWaitlister = mvpm.Type(102)
)

type AccountID = flake.ID

type Account struct {
	ID       AccountID `msgpack:"-"`
	Name     string    `msgpack:"n"`
	Disabled bool      `msgpack:"dis,omitempty"`
}

type AccountObjectKey struct {
	AccountID AccountID
	ObjectID  flake.ID
}

func AccountObject(a AccountID, o flake.ID) AccountObjectKey {
	return AccountObjectKey{a, o}
}

type AccountStringKey struct {
	AccountID AccountID
	String    string
}

func AccountString(a AccountID, s string) AccountStringKey {
	return AccountStringKey{a, s}
}

// type Superadmin struct {
// 	ID        flake.ID `msgpack:"-"`
// 	Email     string   `msgpack:"e"`
// 	EmailNorm string   `msgpack:"e!"`
// 	Name      string   `msgpack:"n"`
// }

// func (obj *Superadmin) FlakeID() flake.ID {
// 	return obj.ID
// }
// func (Superadmin) ObjectType() mvpm.Type {
// 	return TypeSuperadmin
// }

type Waitlister struct {
	ID        flake.ID  `msgpack:"-"`
	Email     string    `msgpack:"e"`
	EmailNorm string    `msgpack:"e!"`
	LastLogin time.Time `msgpack:"@l"`
}

func (obj *Waitlister) FlakeID() flake.ID {
	return obj.ID
}
func (Waitlister) ObjectType() mvpm.Type {
	return TypeWaitlister
}

type Session struct {
	ID                 flake.ID  `msgpack:"-"`
	Actor              mvpm.Ref  `msgpack:"ac"`
	AccountID          flake.ID  `msgpack:"a"`
	ImpersonatedUserID flake.ID  `msgpack:"iu"`
	LastActivity       time.Time `msgpack:"@l"`
	Disabled           bool      `msgpack:"dis"`
}

type UserSignInAttempt struct {
	Email string    `msgpack:"-"`
	Code  string    `msgpack:"c"`
	Time  time.Time `msgpack:"tm"`
}

type Actor interface {
	mvpm.Object
	// ObjectAccountID() flake.ID
}
