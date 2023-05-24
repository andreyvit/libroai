package m

import "github.com/andreyvit/buddyd/mvp/flake"

type ContentID = flake.ID

type Content struct {
	ID        ContentID   `msgpack:"-"`
	AccountID AccountID   `msgpack:"a"`
	ItemID    ItemID      `msgpack:"i"`
	Role      ContentRole `msgpack:"r"`
	Ordinal   int         `msgpack:"o"`
	Text      string      `msgpack:"t,omitempty"`
}

type ContentIROKey struct {
	ItemID  ItemID      `msgpack:"i"`
	Role    ContentRole `msgpack:"r"`
	Ordinal int         `msgpack:"o"`
}

type ContentIOKey struct {
	ItemID  ItemID `msgpack:"i"`
	Ordinal int    `msgpack:"o"`
}

type ContentGroupVM struct {
	Role     ContentRole
	Contents []*Content
}
