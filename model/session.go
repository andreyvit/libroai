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
