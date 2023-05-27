package main

import (
	"fmt"
	"time"

	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/buddyd/mvp/hotwired"
)

func (app *App) handleChatEventStream(rc *mvp.RC, in *struct {
	LastEventID uint64 `form:"Last-Event-ID,header,optional" json:"-"`
}) (any, error) {
	var n uint64 = 1

	if in.LastEventID != 0 {
		n = in.LastEventID + 1
	}

	for rc.Err() == nil {
		rc.SendTurboStream(n, func(stream *hotwired.Stream) {
			stream.Append("message-list", fmt.Sprintf("<div>%d</div>", n))
		})
		n++
		time.Sleep(time.Second)
	}
	return mvp.ResponseHandled{}, nil
}
