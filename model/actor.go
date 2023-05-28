package m

import "github.com/andreyvit/mvp/flake"

type ActorRef struct {
	ActorType ActorType `msgpack:"t"`
	ActorID   flake.ID  `msgpack:"id"`
}
