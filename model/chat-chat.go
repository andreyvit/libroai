package m

import (
	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/openai"
)

type (
	ChatID = flake.ID

	Chat struct {
		ID              ChatID       `msgpack:"-"`
		AccountID       AccountID    `msgpack:"a"`
		UserID          UserID       `msgpack:"u"`
		Cost            openai.Price `msgpack:"c"`
		Title           string       `msgpack:"t,omitempty"`
		TitleCustomized bool         `msgpack:"tc,omitempty"`
		TitleGenerated  bool         `msgpack:"tg,omitempty"`
		TitleRegen      bool         `msgpack:"trg,omitempty"`
	}

	ChatContent struct {
		ChatID ChatID  `msgpack:"-"`
		Turns  []*Turn `msgpack:"t"`
		// LastEventID uint64  `msgpack:"le"`
	}

	ChatVM struct {
		*Chat
		Messages []*MessageVM
	}
)

func (chat *Chat) IsGeneratingTitle() bool {
	return chat.TitleRegen || (!chat.TitleGenerated && !chat.TitleCustomized)
}

func (chat *Chat) TitleWithFallback() string {
	if chat.Title != "" {
		return chat.Title
	}
	return chat.ID.String()
}

func (chat *Chat) NavItemHTMLElementID() string {
	return "chat_nav_" + chat.ID.String()
}

func (cc *ChatContent) LastTurn() *Turn {
	n := len(cc.Turns)
	if n == 0 {
		return nil
	}
	return cc.Turns[n-1]
}

func (cc *ChatContent) FirstUserMessage(beforeTurnIndex int) *Message {
	if t := cc.FirstUserTurn(beforeTurnIndex); t != nil {
		return t.LastMessage()
	}
	return nil
}

func (cc *ChatContent) LastUserMessage(beforeTurnIndex int) *Message {
	if t := cc.LastUserTurn(beforeTurnIndex); t != nil {
		return t.LastMessage()
	}
	return nil
}

func (cc *ChatContent) FirstUserTurn(beforeTurnIndex int) *Turn {
	for i, t := range cc.Turns {
		if beforeTurnIndex >= 0 && i >= beforeTurnIndex {
			break
		}
		if t.Role == MessageRoleUser {
			return t
		}
	}
	return nil
}

func (cc *ChatContent) LastUserTurn(beforeTurnIndex int) *Turn {
	start := len(cc.Turns) - 1
	if beforeTurnIndex >= 0 && start >= beforeTurnIndex {
		start = beforeTurnIndex - 1
	}
	for i := start; i >= 0; i-- {
		t := cc.Turns[i]
		if t.Role == MessageRoleUser {
			return t
		}
	}
	return nil
}

func (cc *ChatContent) FindMessage(msgID MessageID) (*Turn, *Message) {
	for _, t := range cc.Turns {
		if msg := t.Message(msgID); msg != nil {
			return t, msg
		}
	}
	return nil, nil
}

func (cc *ChatContent) FreshMessage(staleMsg *Message) *Message {
	return cc.Message(staleMsg.TurnIndex, staleMsg.ID)
}

func (cc *ChatContent) Message(turnIndex int, msgID MessageID) *Message {
	return cc.Turns[turnIndex].Message(msgID)
}

func WrapChat(chat *Chat, content *ChatContent) *ChatVM {
	chatVM := &ChatVM{
		Chat:     chat,
		Messages: make([]*MessageVM, 0, len(content.Turns)),
	}
	for _, t := range content.Turns {
		msg := t.LastMessage()
		chatVM.Messages = append(chatVM.Messages, WrapMessage(msg, chat.ID))
	}
	return chatVM
}
