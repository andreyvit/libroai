package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
)

func loadAllChatListMiddleware(rc *RC) (any, error) {
	rc.Chats = edb.All(edb.IndexScan[m.Chat](rc, ChatsByAccount, edb.ExactScan(rc.AccountID()).Reversed()))
	return nil, nil
}
