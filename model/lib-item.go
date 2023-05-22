package m

import "github.com/andreyvit/buddyd/mvp/flake"

type ItemID = flake.ID

type Item struct {
	ID        ItemID    `msgpack:"-"`
	AccountID AccountID `msgpack:"a"`
	FolderID  FolderID  `msgpack:"f"`
	Name      string    `msgpack:"n"`
	FileName  string    `msgpack:"fn,omitempty"`
	Link      string    `msgpack:"l,omitempty"`
}
