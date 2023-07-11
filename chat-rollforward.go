package main

import (
	"context"
	"fmt"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/multierr"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/flogger"
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
	var needTitle bool
	err := app.InTx(&rc.RC, mvpm.SafeReader, func() error {
		chat := edb.Get[m.Chat](rc, chatID)
		cc := edb.Get[m.ChatContent](rc, chatID)
		unembeddedMsgs = findMessagesWithMissingEmbeddings(cc)
		pendingBotMsg = findPendingBotMessage(cc)
		needTitle = chat.IsGeneratingTitle()
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

	if pendingBotMsg == nil && !needTitle {
		return nil
	}

	if pendingBotMsg != nil {
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

		newBotMsg, newBotMsgErr := openai.StreamChat(rc, history, opt, app.httpClient, app.Settings().OpenAICreds, func(msg *openai.Msg, delta string) error {
			flogger.Log(rc, "openai chunk: <<<%s>>>", delta)
			pendingBotMsg.Text = msg.Content
			pushMessage(rc, chatID, pendingBotMsg)
			return nil
		})

		spent := openai.Cost(openai.ChatTokenCount(history, opt.Model), openai.MsgTokenCount(newBotMsg, opt.Model), opt.Model)

		err = app.InTx(&rc.RC, mvpm.SafeWriter, func() error {
			chat := edb.Get[m.Chat](rc, chatID)
			cc := edb.Get[m.ChatContent](rc, chatID)
			msg := cc.FreshMessage(pendingBotMsg)

			chat.Cost += spent

			if msg == nil {
				flogger.Log(rc, "WARNING: bot message not found for pendingBotMsg %v %v", pendingBotMsg.ID, pendingBotMsg)
				pendingBotMsg = nil
			} else {
				if newBotMsgErr != nil {
					msg.State = m.MessageStateFailed
				} else {
					msg.Text = newBotMsg.Content
					msg.State = m.MessageStateFinished
				}
				pendingBotMsg = msg
			}
			edb.Put(rc, chat, cc)
			return nil
		})
		if err != nil {
			return err
		}
		if pendingBotMsg != nil {
			pushMessage(rc, chatID, pendingBotMsg)
		}
	}

	if needTitle {
		var history []openai.Msg
		history = append(history, openai.SystemMsg(chatTitleSystemPrompt))
		err = app.InTx(&rc.RC, mvpm.SafeReader, func() error {
			cc := edb.Get[m.ChatContent](rc, chatID)
			for _, t := range cc.Turns {
				if t.Role == m.MessageRoleUser {
					msg := t.LastMessage()
					history = append(history, openai.Msg{
						Role:    msg.Role.OpenAIRole(),
						Content: msg.Text,
					})
				}
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
		opt.Functions = []any{chatTitleFunc}
		opt.FunctionCallMode = &openai.ForceFunctionCall{Name: chatTitleFuncName}

		newTitleMsgs, usage, err := openai.Chat(rc, history, opt, app.httpClient, app.Settings().OpenAICreds)
		spent := openai.Cost(usage.PromptTokens, usage.CompletionTokens, opt.Model)
		var newTitle string
		if err == nil {
			var result ChatTitleFuncResult
			err = newTitleMsgs[0].UnmarshalCallArguments(&result)
			if err == nil {
				newTitle = result.Title
				if newTitle == "" {
					err = fmt.Errorf("ChatGPT returned an empty title")
				}
			}
		}

		if err != nil {
			flogger.Log(rc, "WARNING: title generation failed: %v", err)
		}

		var chat *m.Chat
		err = app.InTx(&rc.RC, mvpm.SafeWriter, func() error {
			chat = edb.Get[m.Chat](rc, chatID)

			chat.Cost += spent

			if needTitle && (chat.TitleRegen || !chat.TitleCustomized) {
				if newTitle != "" {
					chat.TitleGenerated = true
				}
				if newTitle == "" && chat.Title == "" {
					newTitle = chat.ID.Time().Format("Chat Jan 02")
				}
				if newTitle != "" {
					chat.Title = newTitle
					if false {
						chat.Title += time.Now().Format(" 15:04:05")
					}
				}
				chat.TitleRegen = false
			}

			edb.Put(rc, chat)
			return nil
		})
		if err != nil {
			return err
		}

		pushChatTitle(rc, chat)
	}

	return nil
}

func pushMessage(rc *RC, chatID m.ChatID, msg *m.Message) {
	mvp.PushPartial(rc, &mvp.ViewData{
		View: "chat/_message",
		Data: m.WrapMessage(msg, chatID),
	}, msg.HTMLElementID(), chatChannel(chatID), mvplive.Envelope{
		DedupKey: msg.ID.String(),
	})
}

func pushChatTitle(rc *RC, chat *m.Chat) {
	mvp.PushPartial(rc, &mvp.ViewData{
		View:         "chat/_nav_item",
		Data:         chat,
		SemanticPath: chat.UserChatSempath(),
	}, chat.UserNavItemHTMLElementID(), chatChannel(chat.ID), mvplive.Envelope{
		DedupKey: "title",
	})
	mvp.PushPartial(rc, &mvp.ViewData{
		View:         "chat/_nav_item_mod",
		Data:         chat,
		SemanticPath: chat.ModChatSempath(),
	}, chat.ModNavItemHTMLElementID(), chatChannel(chat.ID), mvplive.Envelope{
		DedupKey: "title-mod",
	})
}

func chatChannel(chatID m.ChatID) mvplive.Channel {
	return mvplive.Channel{
		Family: chatChannelFamily,
		Topic:  chatID.String(),
	}
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
