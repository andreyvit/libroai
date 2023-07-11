package main

import (
	"fmt"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/httperrors"
	"github.com/andreyvit/openai"

	m "github.com/andreyvit/buddyd/model"
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
	content := loadChatContent(rc, chat.ID)
	return &mvp.ViewData{
		View:         "chat/chat",
		Title:        "Chat",
		SemanticPath: chat.UserChatSempath(),
		Data: struct {
			IsModerator bool
			IsNewChat   bool
			Chat        *m.ChatVM
		}{
			IsModerator: false,
			IsNewChat:   chat.ID == 0,
			Chat:        m.WrapChat(chat, content),
		},
	}, nil
}

func (app *App) showModChat(rc *RC, in *struct {
	ChatID flake.ID `form:"chat,path" json:"-"`
}) (*mvp.ViewData, error) {
	chat, err := loadChat(rc, in.ChatID, false)
	if err != nil {
		return nil, err
	}
	content := loadChatContent(rc, chat.ID)
	return &mvp.ViewData{
		View:         "chat/chat",
		Title:        "Chat",
		SemanticPath: chat.ModChatSempath(),
		Data: struct {
			IsModerator bool
			IsNewChat   bool
			Chat        *m.ChatVM
		}{
			IsModerator: true,
			IsNewChat:   false,
			Chat:        m.WrapChat(chat, content),
		},
	}, nil
}

func (app *App) sendChatMessage(rc *RC, in *struct {
	ChatID  flake.ID `form:"chat,path" json:"-"`
	Message string   `json:"message"`
}) (any, error) {
	if in.Message == "" {
		return app.Redirect("chat.view", ":chat", in.ChatID), nil
	}
	if openai.TokenCount(in.Message, DefaultModel) > MaxMsgTokenCount {
		return nil, fmt.Errorf("message too long")
	}

	chat, err := loadChat(rc, in.ChatID, true)
	if err != nil {
		return nil, err
	}
	cc := loadChatContent(rc, chat.ID)
	if chat.ID == 0 {
		chat.ID = app.NewID()
		cc.ChatID = chat.ID
	}

	userTurn := app.addTurn(cc, m.MessageRoleUser)
	app.addUserMsg(userTurn, in.Message)

	botTurn := app.addTurn(cc, m.MessageRoleBot)
	app.addBotPendingMsg(botTurn)

	edb.Put(rc, chat, cc)
	app.EnqueueChatRollforward(rc, chat.ID)

	return app.Redirect("chat.view", ":chat", chat.ID), nil
}

func (app *App) markChatMessage(rc *RC, in *struct {
	ChatID    flake.ID `form:"chat,path" json:"-"`
	MessageID flake.ID `form:"message,path" json:"-"`
	Action    string   `json:"action"`
}) (any, error) {
	chat := must(loadChat(rc, in.ChatID, false))
	cc := loadChatContent(rc, chat.ID)
	turn, msg := cc.FindMessage(in.MessageID)
	if turn == nil {
		return nil, httperrors.NotFound
	}
	if turn.Role != m.MessageRoleBot {
		return nil, httperrors.BadRequest.Msg("cannot mark or regen user messages")
	}

	var rollforward bool
	switch in.Action {
	case "regen":
		if !turn.IsLastMessagePending() {
			app.addBotPendingMsg(turn)
		}
		rollforward = true
	case "voteup":
		msg.VotedUp = true
		msg.VotedDown = false
	case "undo-voteup":
		msg.VotedUp = false
	case "votedown":
		msg.VotedDown = true
		msg.VotedUp = false
	case "undo-votedown":
		msg.VotedDown = false
	default:
		return nil, httperrors.BadRequest.Msg("invalid action")
	}

	edb.Put(rc, chat, cc)
	if rollforward {
		app.EnqueueChatRollforward(rc, chat.ID)
	}

	return app.Redirect("chat.view", ":chat", chat.ID), nil
}

func (app *App) handleChatAction(rc *RC, in *struct {
	ChatID flake.ID `form:"chat,path" json:"-"`
	Action string   `form:"action,path" json:"-"`
}) (any, error) {
	chat := must(loadChat(rc, in.ChatID, false))
	cc := loadChatContent(rc, chat.ID)

	var rollforward bool
	switch in.Action {
	case "retitle":

	default:
		return nil, httperrors.BadRequest.Msg("invalid action")
	}

	edb.Put(rc, chat, cc)
	if rollforward {
		app.EnqueueChatRollforward(rc, chat.ID)
	}

	return rc.RedirectBack(), nil
}

func (app *App) doRetitleChat(rc *RC, in *struct {
	ChatID flake.ID `json:"chat_id"`
}) (any, error) {
	chat := must(loadChat(rc, in.ChatID, false))
	cc := loadChatContent(rc, chat.ID)

	chat.TitleRegen = true

	edb.Put(rc, chat, cc)
	app.EnqueueChatRollforward(rc, chat.ID)
	return rc.RedirectBack(), nil
}
