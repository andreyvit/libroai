package m

import (
	"time"

	"github.com/andreyvit/buddyd/internal/flake"
)

type Session struct {
	ID           flake.ID  `msgpack:"-"`
	Actor        ActorRef  `msgpack:"a"`
	CreationTime time.Time `msgpack:"tmc"`
	RefreshTime  time.Time `msgpack:"tmr"`
	ActivityTime time.Time `msgpack:"tma"`
}

type UserSignInAttempt struct {
	Email string    `msgpack:"-"`
	Code  string    `msgpack:"c"`
	Time  time.Time `msgpack:"tm"`
}
