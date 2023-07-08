package m

import (
	"strings"

	"github.com/andreyvit/mvp/flake"
)

type MessageID = flake.ID

type Message struct {
	ID                MessageID    `msgpack:"#"`
	Role              MessageRole  `msgpack:"r"`
	State             MessageState `msgpack:"s"`
	Text              string       `msgpack:"t"`
	TurnID            TurnID       `msgpack:"tid"`
	TurnIndex         int          `msgpack:"ti"`
	EmbeddingAda002   Embedding    `msgpack:"e2,omitempty"`
	ContextContentIDs []ContentID  `msgpack:"cc,omitempty"`
	ContextDistances  []float64    `msgpack:"cd,omitempty"`

	VotedUp   bool `msgpack:"vu,omitempty"`
	VotedDown bool `msgpack:"vd,omitempty"`
}

func (msg *Message) IDOrZero() MessageID {
	if msg == nil {
		return 0
	}
	return msg.ID
}

func (msg *Message) Voted() bool {
	return msg.VotedUp || msg.VotedDown
}

func (msg *Message) HTMLElementID() string {
	return "message_" + msg.ID.String()
}

type MessageVM struct {
	*Message
	ChatID      ChatID
	IsVotedUp   bool
	IsVotedDown bool
}

func (m *MessageVM) Paragraphs() []string {
	return strings.FieldsFunc(m.Text, isNewLine)
}

func WrapMessage(msg *Message, chatID ChatID) *MessageVM {
	return &MessageVM{
		Message: msg,
		ChatID:  chatID,
	}
}
