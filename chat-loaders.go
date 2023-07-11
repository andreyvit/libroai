package main

import (
	"github.com/andreyvit/edb"
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
	loadRuntimeAccount(rc, rc.AccountID())
	rc.Chats = wrapChatList(rc, edb.All(edb.ReverseExactIndexScan[m.Chat](rc, ChatsByAccountUser, m.AccountUser(rc.AccountID(), rc.UserID()))))
	return nil, nil
}

func wrapChatList(rc *RC, rawChats []*m.Chat) []*m.ChatVM {
	chats := make([]*m.ChatVM, 0, len(rawChats))
	for _, rawChat := range rawChats {
		chat := &m.ChatVM{
			Chat:   rawChat,
			Author: rc.Account.UserByID(rawChat.UserID),
		}
		chats = append(chats, chat)
	}
	return chats
}
