package m

import "github.com/andreyvit/buddyd/mvp/flake"

type ActorRef struct {
	ActorType ActorType `msgpack:"t"`
	ActorID   flake.ID  `msgpack:"id"`
}
