package m

import "github.com/andreyvit/mvp/flake"

type TurnID = flake.ID

type Turn struct {
	ID       TurnID      `msgpack:"i"`
	Index    int         `msgpack:"ti"`
	Role     MessageRole `msgpack:"r"`
	Versions []*Message  `msgpack:"m"`
}

func (t *Turn) IsLastMessagePending() bool {
	if msg := t.LastMessage(); msg != nil {
		return msg.State == MessageStatePending
	}
	return false
}

func (t *Turn) LastMessage() *Message {
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
