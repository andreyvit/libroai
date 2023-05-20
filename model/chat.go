package m

import (
	"strings"

	"github.com/andreyvit/buddyd/mvp/flake"
	"github.com/andreyvit/openai"
)

type ChatID = flake.ID

type Chat struct {
	ID        ChatID       `msgpack:"-"`
	AccountID AccountID    `msgpack:"a"`
	UserID    UserID       `msgpack:"u"`
	Cost      openai.Price `msgpack:"c"`
}

type ChatContent struct {
	ChatID   ChatID     `msgpack:"-"`
	Messages []*Message `msgpack:"m"`
}

type MessageID = flake.ID

type Message struct {
	ID   MessageID   `msgpack:"#"`
	Role MessageRole `msgpack:"r"`
	Text string      `msgpack:"t"`
}

type ChatVM struct {
	*Chat
	Messages []*MessageVM
}

type MessageVM struct {
	*Message
}

func (m *MessageVM) Paragraphs() []string {
	return strings.FieldsFunc(m.Text, isNewLine)
}

func WrapChat(chat *Chat, content *ChatContent) *ChatVM {
	chatVM := &ChatVM{
		Chat:     chat,
		Messages: make([]*MessageVM, 0, len(content.Messages)),
	}
	for _, msg := range content.Messages {
		chatVM.Messages = append(chatVM.Messages, &MessageVM{
			Message: msg,
		})
	}
	return chatVM
}
