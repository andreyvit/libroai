package main

import (
	"context"
	"fmt"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/multierr"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/hotwired"
	"github.com/andreyvit/mvp/mvplive"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"github.com/andreyvit/openai"

	m "github.com/andreyvit/buddyd/model"
)

func (app *App) EnqueueChatRollforward(rc *RC, chatID m.ChatID) {
	app.EnqueueEphemeral(jobProduceAnswer, chatID.String(), func(rc *mvp.RC) error {
		return app.runChatRollforward(fullRC.From(rc), chatID)
	})
}

func (app *App) runChatRollforward(rc *RC, chatID m.ChatID) error {
	var unembeddedMsgs []*m.Message
	var pendingBotMsg *m.Message
	err := app.InTx(&rc.RC, mvpm.SafeReader, func() error {
		cc := edb.Get[m.ChatContent](rc, chatID)
		unembeddedMsgs = findMessagesWithMissingEmbeddings(cc)
		pendingBotMsg = findPendingBotMessage(cc)
		return nil
	})

	flogger.Log(rc, "ChatRollforward(%v): unembeddedMsgs=%d pendingBotMsg=%v", chatID, len(unembeddedMsgs), pendingBotMsg.IDOrZero())

	if len(unembeddedMsgs) > 0 {
		var embeddingErr error
		var embeddingCost openai.Price
		for _, msg := range unembeddedMsgs {
			cost, err := app.computeMsgEmbedding(rc, msg)
			embeddingCost += cost
			embeddingErr = multierr.Append(embeddingErr, err)
		}

		// save in a separate transaction to avoid blocking writes
		// while computing embedding matches below
		err := app.InTx(&rc.RC, mvpm.SafeWriter, func() error {
			chat := edb.Get[m.Chat](rc, chatID)
			cc := edb.Get[m.ChatContent](rc, chatID)
			chat.Cost += embeddingCost
			for _, msg := range unembeddedMsgs {
				if newMsg := cc.FreshMessage(msg); newMsg != nil {
					newMsg.EmbeddingAda002 = msg.EmbeddingAda002
				}
			}
			edb.Put(rc, chat, cc)
			return nil
		})
		if err != nil {
			return err
		}
	}

	if pendingBotMsg == nil {
		return nil
	}

	var history []openai.Msg
	err = app.InTx(&rc.RC, mvpm.SafeReader, func() error {
		chat := edb.Get[m.Chat](rc, chatID)
		cc := edb.Get[m.ChatContent](rc, chatID)
		embs := loadAccountEmbeddings(rc, chat.AccountID)

		pres, err := app.BuildSystemPrompt(rc, prompt1, cc, pendingBotMsg.TurnIndex, embs)
		if err != nil {
			return err
		}

		flogger.Log(rc, "Prompt: %s", pres.Prompt)

		history = append(history, openai.SystemMsg(pres.Prompt))
		for i, t := range cc.Turns {
			if i >= pendingBotMsg.TurnIndex {
				break
			}
			msg := t.LastMessage()
			history = append(history, openai.Msg{
				Role:    msg.Role.OpenAIRole(),
				Content: msg.Text,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	opt := openai.DefaultChatOptions()
	opt.Model = DefaultModel
	opt.MaxTokens = MaxResponseTokenCount
	opt.Temperature = 0.75

	botMsg, chatErr := openai.StreamChat(rc, history, opt, app.httpClient, app.Settings().OpenAICreds, func(msg *openai.Msg, delta string) error {
		flogger.Log(rc, "openai chunk: <<<%s>>>", delta)
		pendingBotMsg.Text = msg.Content
		pushMessage(&rc.RC, chatID, pendingBotMsg)
		return nil
	})

	err = app.InTx(&rc.RC, mvpm.SafeWriter, func() error {
		chat := edb.Get[m.Chat](rc, chatID)
		cc := edb.Get[m.ChatContent](rc, chatID)
		msg := cc.FreshMessage(pendingBotMsg)

		chat.Cost += openai.Cost(openai.ChatTokenCount(history, opt.Model), openai.MsgTokenCount(botMsg, opt.Model), opt.Model)

		if chatErr != nil {
			msg.State = m.MessageStateFailed
		} else {
			msg.Text = botMsg.Content
			msg.State = m.MessageStateFinished
		}
		pendingBotMsg = msg
		edb.Put(rc, chat, cc)
		return nil
	})
	if err != nil {
		return err
	}

	pushMessage(&rc.RC, chatID, pendingBotMsg)

	return nil
}

func pushMessage(rc *mvp.RC, chatID m.ChatID, msg *m.Message) {
	app := rc.App()
	content := app.RenderPartial(rc, "chat/_message", m.WrapMessage(msg, chatID))
	app.PublishTurbo(rc, mvplive.Channel{
		Family: chatChannelFamily,
		Topic:  chatID.String(),
	}, mvplive.Envelope{
		DedupKey: msg.ID.String(),
	}, func(stream *hotwired.Stream) {
		stream.Replace(msg.HTMLElementID(), string(content))
	})
}

func findMessagesWithMissingEmbeddings(cc *m.ChatContent) []*m.Message {
	var unembeddedMsgs []*m.Message
	for _, turn := range cc.Turns {
		if turn.Role == m.MessageRoleUser {
			for _, msg := range turn.Versions {
				if msg.EmbeddingAda002 == nil {
					unembeddedMsgs = append(unembeddedMsgs, msg)
				}
			}
		}
	}
	return unembeddedMsgs
}

func findPendingBotMessage(cc *m.ChatContent) *m.Message {
	for i := len(cc.Turns) - 1; i >= 0; i-- {
		turn := cc.Turns[i]
		if turn.Role == m.MessageRoleBot {
			if msg := turn.LastMessage(); msg != nil && msg.State == m.MessageStatePending {
				return msg
			}
		}
	}
	return nil
}

func (app *App) computeMsgEmbedding(ctx context.Context, msg *m.Message) (openai.Price, error) {
	embedding, usage, err := openai.ComputeEmbedding(ctx, msg.Text, app.httpClient, app.Settings().OpenAICreds)
	if err != nil {
		return 0, fmt.Errorf("embeddings: %w", err)
	}
	msg.EmbeddingAda002 = embedding
	cost := openai.Cost(usage.PromptTokens, usage.CompletionTokens, EmbeddingModel)
	return cost, nil
}
