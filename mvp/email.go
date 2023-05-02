package mvp

import (
	"github.com/andreyvit/buddyd/internal/postmark"
)

type Email struct {
	From     string
	To       string
	ReplyTo  string
	Cc       string
	Bcc      string
	Subject  string
	TextBody string
	HtmlBody string

	TemplateIntID int64
	TemplateData  map[string]any

	MessageStream string
	Category      string
}

func (app *App) SendEmail(rc *RC, msg *Email) {
	pmsg := &postmark.Message{
		From:          msg.From,
		To:            msg.To,
		ReplyTo:       msg.ReplyTo,
		Cc:            msg.Cc,
		Bcc:           msg.Bcc,
		Subject:       msg.Subject,
		Tag:           msg.Category,
		TextBody:      msg.TextBody,
		HtmlBody:      msg.HtmlBody,
		MessageStream: msg.MessageStream,
		TemplateId:    msg.TemplateIntID,
		TemplateModel: msg.TemplateData,
	}
	app.postmrk.Send(rc, pmsg)
}
