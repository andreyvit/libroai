package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/buddyd/mvp/flake"
	"github.com/andreyvit/buddyd/mvp/mvplive"
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
