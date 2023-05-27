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
	ChatID ChatID  `msgpack:"-"`
	Turns  []*Turn `msgpack:"t"`
}

func (cc *ChatContent) FirstUserMessage() *Message {
	if t := cc.FirstUserTurn(); t != nil {
		return t.LatestVersion()
	}
	return nil
}

func (cc *ChatContent) LatestUserMessage() *Message {
	if t := cc.LatestUserTurn(); t != nil {
		return t.LatestVersion()
	}
	return nil
}

func (cc *ChatContent) FirstUserTurn() *Turn {
	for _, t := range cc.Turns {
		if t.Role == MessageRoleUser {
			return t
		}
	}
	return nil
}

func (cc *ChatContent) LatestUserTurn() *Turn {
	n := len(cc.Turns)
	for i := n - 1; i >= 0; i-- {
		t := cc.Turns[i]
		if t.Role == MessageRoleUser {
			return t
		}
	}
	return nil
}

func (cc *ChatContent) Message(turnIndex int, msgID MessageID) *Message {
	return cc.Turns[turnIndex].Message(msgID)
}

type Turn struct {
	Role     MessageRole `msgpack:"r"`
	Versions []*Message  `msgpack:"m"`
}

func (t *Turn) LatestVersion() *Message {
	return t.Versions[len(t.Versions)-1]
}

func (t *Turn) Message(msgID MessageID) *Message {
	for _, msg := range t.Versions {
		if msg.ID == msgID {
			return msg
		}
	}
	return nil
}

type MessageID = flake.ID

type Message struct {
	ID                MessageID    `msgpack:"#"`
	Role              MessageRole  `msgpack:"r"`
	State             MessageState `msgpack:"s"`
	Text              string       `msgpack:"t"`
	EmbeddingAda002   Embedding    `msgpack:"e2,omitempty"`
	ContextContentIDs []ContentID  `msgpack:"cc,omitempty"`
	ContextDistances  []float64    `msgpack:"cd,omitempty"`
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
		Messages: make([]*MessageVM, 0, len(content.Turns)),
	}
	for _, t := range content.Turns {
		msg := t.LatestVersion()
		chatVM.Messages = append(chatVM.Messages, &MessageVM{
			Message: msg,
		})
	}
	return chatVM
}
