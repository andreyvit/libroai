package m

import "github.com/andreyvit/buddyd/internal/flake"

type ActorRef struct {
	ActorType ActorType `msgpack:"t"`
	ActorID   flake.ID  `msgpack:"id"`
}
