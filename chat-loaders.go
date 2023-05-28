package main

import (
	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/httperrors"

	m "github.com/andreyvit/buddyd/model"
)

func loadChat(rc *RC, chatID m.ChatID, allowNew bool) (*m.Chat, error) {
	if chatID == 0 {
		if allowNew {
			return &m.Chat{
				AccountID: rc.AccountID(),
				UserID:    rc.UserID(),
			}, nil
		} else {
			return nil, httperrors.Errorf(400, "", "Invalid chat ID")
		}
	}
	chat := edb.Get[m.Chat](rc, chatID)
	if chat == nil || chat.AccountID != rc.AccountID() || chat.UserID != rc.UserID() {
		return nil, httperrors.Errorf(404, "chat_not_found", "This chat does not exist.")
	}
	return chat, nil
}

func loadChatContent(rc *RC, chatID m.ChatID) *m.ChatContent {
	if chatID == 0 {
		return &m.ChatContent{}
	} else {
		result := edb.Get[m.ChatContent](rc, chatID)
		if result == nil {
			result = &m.ChatContent{
				ChatID: chatID,
			}
		}
		return result
	}
}

func loadUserChatListMiddleware(rc *RC) (any, error) {
	rc.Chats = edb.All(edb.IndexScan[m.Chat](rc, ChatsByUser, edb.ExactScan(rc.UserID()).Reversed()))
	flogger.Log(rc, "rc.Chats = %v", rc.Chats)
	return nil, nil
}
