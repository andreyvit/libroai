package main

import (
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/mvplive"

	m "github.com/andreyvit/buddyd/model"
)

var (
	chatChannelFamily = &mvplive.ChannelFamily{
		Name: "chat",
	}
)

func (app *App) handleChatEventStream(rc *mvp.RC, in *struct {
	ChatID      m.ChatID `form:"chat,path" json:"-"`
	LastEventID uint64   `form:"Last-Event-ID,header,optional" json:"-"`
}) (any, error) {
	app.Subscribe(rc, rc, rc.RespWriter, mvplive.Channel{
		Family: chatChannelFamily,
		Topic:  in.ChatID.String(),
	}, flake.ID(in.LastEventID))
	return mvp.ResponseHandled{}, nil
}
