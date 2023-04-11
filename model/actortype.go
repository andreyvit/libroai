package m

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type ActorType int

const (
	ActorTypeAnonymous = ActorType(iota)
	ActorTypeUser
	ActorTypeAdmin
	ActorTypeTenant
	ActorTypeCustomer
)

var (
	_actorTypes          = []string{"anonymous", "user", "admin", "tenant", "customer"}
	_actorTypeKeyStrings = []string{"", "", "A", "T", "C"}
)

func (v ActorType) String() string {
	return _actorTypes[v]
}

func ParseActorType(s string) (ActorType, error) {
	if i := slices.Index(_actorTypes, s); i >= 0 {
		return ActorType(i), nil
	} else {
		return ActorTypeAnonymous, fmt.Errorf("invalid ActorType %q", s)
	}
}

func (v ActorType) KeyString() string {
	return _actorTypeKeyStrings[v]
}

func ParseActorTypeByKeyString(s string) (ActorType, error) {
	if i := slices.Index(_actorTypeKeyStrings, s); i >= 0 {
		return ActorType(i), nil
	} else {
		return ActorTypeAnonymous, fmt.Errorf("invalid ActorType %q", s)
	}
}

func (v ActorType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *ActorType) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseActorType(string(b))
	return err
}
