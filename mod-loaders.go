package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
)

func loadAllChatListMiddleware(rc *RC) (any, error) {
	rc.Chats = wrapChatList(rc, edb.All(edb.ReverseExactIndexScan[m.Chat](rc, ChatsByAccount, rc.AccountID())))
	return nil, nil
}
