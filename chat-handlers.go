package main

import (
	"fmt"

	"github.com/andreyvit/buddyd/internal/flogger"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/buddyd/mvp/flake"
	"github.com/andreyvit/edb"
	"github.com/andreyvit/openai"
)

func (app *App) showNewChat(rc *RC, in *struct{}) (*mvp.ViewData, error) {
	chat := &m.Chat{
		AccountID: rc.AccountID(),
		UserID:    rc.UserID(),
	}
	return app.doShowChat(rc, chat)
}

func (app *App) showChat(rc *RC, in *struct {
	ChatID flake.ID `form:"chat,path" json:"-"`
}) (*mvp.ViewData, error) {
	chat, err := loadChat(rc, in.ChatID, false)
	if err != nil {
		return nil, err
	}
	return app.doShowChat(rc, chat)
}

func (app *App) doShowChat(rc *RC, chat *m.Chat) (*mvp.ViewData, error) {
	flogger.Log(rc, "rc.Chats x2 = %v", rc.Chats)
	content := loadChatContent(rc, chat.ID)
	return &mvp.ViewData{
		View:         "chat/chat",
		Title:        "Chat",
		SemanticPath: fmt.Sprintf("chat/c/%v", chat.ID),
		Data: struct {
			IsNewChat bool
			Chat      *m.ChatVM
		}{
			IsNewChat: chat.ID == 0,
			Chat:      m.WrapChat(chat, content),
		},
	}, nil
}

func (app *App) sendChatMessage(rc *RC, in *struct {
	ChatID  flake.ID `form:"chat,path" json:"-"`
	Message string   `json:"message"`
}) (any, error) {
	if in.Message == "" {
		return app.Redirect("chat.view", "chat", in.ChatID), nil
	}
	if openai.TokenCount(in.Message, DefaultModel) > MaxMsgTokenCount {
		return nil, fmt.Errorf("message too long")
	}

	chat, err := loadChat(rc, in.ChatID, true)
	if err != nil {
		return nil, err
	}
	content := loadChatContent(rc, chat.ID)
	isNew := (chat.ID == 0)
	if isNew {
		chat.ID = app.NewID()
		content.ChatID = chat.ID
	}

	content.Messages = append(content.Messages, &m.Message{
		ID:   app.NewID(),
		Role: m.MessageRoleUser,
		Text: in.Message,
	})

	_, err = app.ProduceBotMessage(rc, chat, content)
	if err != nil {
		return nil, err
	}

	edb.Put(rc, chat)
	edb.Put(rc, content)

	// app.SendMessage(rc.Ctx, chat, in.Message)

	return app.Redirect("chat.view", "chat", chat.ID), nil
}

func (app *App) voteChatResponse(rc *RC, in *struct {
	ChatID       flake.ID `form:"chat,path" json:"-"`
	RoundOrdinal int      `form:"round"`
	ResponseKey  string   `form:"key"`
	Vote         string   `form:"vote"`
}) (any, error) {
	chat, err := loadChat(rc, in.ChatID, false)
	if err != nil {
		return nil, err
	}

	_ = chat
	// round := chat.Rounds[in.RoundOrdinal-1]
	// switch in.Vote {
	// case "up":
	// 	round.BestResponseKey = in.ResponseKey
	// case "undo-up":
	// 	if round.BestResponseKey == in.ResponseKey {
	// 		round.BestResponseKey = ""
	// 	}
	// }
	// app.State.SaveChat(chat)

	return app.Redirect("chat.view", "chat", in.ChatID), nil
}
