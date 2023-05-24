package m

import (
	"fmt"

	"github.com/andreyvit/buddyd/mvp/flake"
)

type ItemID = flake.ID

type Item struct {
	ID               ItemID    `msgpack:"-"`
	AccountID        AccountID `msgpack:"a"`
	FolderID         FolderID  `msgpack:"f"`
	Name             string    `msgpack:"n"`
	FileName         string    `msgpack:"fn,omitempty"`
	ImportSourceName string    `msgpack:"isn,omitempty"`
	Link             string    `msgpack:"l,omitempty"`
}

func (item *Item) SemanticPath() string {
	return fmt.Sprintf("lib/items/%v", item.ID)
}
