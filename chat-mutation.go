package main

import m "github.com/andreyvit/buddyd/model"

func (app *App) addTurn(cc *m.ChatContent, role m.MessageRole) *m.Turn {
	turn := &m.Turn{
		ID:       app.NewID(),
		Index:    len(cc.Turns),
		Role:     role,
		Versions: []*m.Message{},
	}
	cc.Turns = append(cc.Turns, turn)
	return turn
}

func (app *App) addUserMsg(turn *m.Turn, content string) *m.Message {
	if turn.Role != m.MessageRoleUser {
		panic("cannot add user msg to bot turn")
	}
	msg := &m.Message{
		ID:        app.NewID(),
		Role:      m.MessageRoleUser,
		Text:      content,
		TurnID:    turn.ID,
		TurnIndex: turn.Index,
	}
	turn.Versions = append(turn.Versions, msg)
	return msg
}

func (app *App) addBotPendingMsg(turn *m.Turn) *m.Message {
	if turn.Role != m.MessageRoleBot {
		panic("cannot add bot msg to user turn")
	}
	msg := &m.Message{
		ID:        app.NewID(),
		Role:      m.MessageRoleBot,
		State:     m.MessageStatePending,
		TurnID:    turn.ID,
		TurnIndex: turn.Index,
	}
	turn.Versions = append(turn.Versions, msg)
	return msg
}
