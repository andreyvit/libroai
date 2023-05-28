package main

import (
	"fmt"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/hotwired"
	"github.com/andreyvit/mvp/mvplive"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"github.com/andreyvit/openai"

	m "github.com/andreyvit/buddyd/model"
)

func (app *App) ProduceBotMessage(rc *RC, chat *m.Chat, cc *m.ChatContent) (*m.Message, error) {
	// embedding, usage, err := openai.ComputeEmbedding(context.Background(), msgContent, app.httpClient, app.OpenAICreds)
	// if err != nil {
	// 	return nil, fmt.Errorf("embeddings: %w", err)
	// }
	// round.UserMessageEmbedding = embedding
	// chat.Rounds = append(chat.Rounds, round)
	// chat.Cost += openai.Cost(usage.PromptTokens, usage.CompletionTokens, EmbeddingModel)
	// app.State.SaveChat(chat)

	// contextEntries, contextRanks := app.memory.Select(embedding)

	turnIndex := len(cc.Turns)
	msg := &m.Message{
		ID:    app.NewID(),
		Role:  m.MessageRoleBot,
		State: m.MessageStatePending,
	}
	cc.Turns = append(cc.Turns, &m.Turn{
		Role: m.MessageRoleBot,
		Versions: []*m.Message{
			msg,
		},
	})

	chatID, msgID := chat.ID, msg.ID

	app.EnqueueEphemeral(jobProduceAnswer, chat.ID.String(), func(rc *mvp.RC) error {
		return app.produceBotMessage(fullRC.From(rc), chatID, turnIndex, msgID)
	})

	return msg, nil
}

func (app *App) produceBotMessage(rc *RC, chatID m.ChatID, turnIndex int, msgID m.MessageID) error {
	var history []openai.Msg
	err := app.InTx(&rc.RC, mvpm.SafeReader, func() error {
		chat := edb.Get[m.Chat](rc, chatID)
		cc := edb.Get[m.ChatContent](rc, chatID)

		loadAccountEmbeddings(rc, false)
		pr, err := app.BuildSystemPrompt(rc, prompt1, chat, cc)
		if err != nil {
			return err
		}

		flogger.Log(rc, "Prompt: %s", pr.Prompt)

		history = append(history, openai.SystemMsg(pr.Prompt))
		for _, t := range cc.Turns {
			msg := t.LatestVersion()
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

	choices, usage, chatErr := openai.Chat(rc, history, opt, app.httpClient, app.Settings().OpenAICreds)
	if chatErr == nil && len(choices) != 1 {
		chatErr = fmt.Errorf("len(choices) = %d", len(choices))
	}

	err = app.InTx(&rc.RC, mvpm.SafeWriter, func() error {
		chat := edb.Get[m.Chat](rc, chatID)
		cc := edb.Get[m.ChatContent](rc, chatID)
		msg := cc.Message(turnIndex, msgID)

		chat.Cost += openai.Cost(usage.PromptTokens, usage.CompletionTokens, opt.Model)

		if chatErr != nil {
			msg.State = m.MessageStateFailed
		} else {
			choice := choices[0]
			msg.Text = choice.Content
			msg.State = m.MessageStateFinished
		}
		cc.LastEventID++
		eventID := cc.LastEventID
		edb.Put(rc, chat)
		edb.Put(rc, cc)

		content := app.RenderPartial(&rc.RC, "chat/_message", m.WrapMessage(msg, chatID, turnIndex))
		app.PublishTurbo(mvplive.Channel{
			Family: chatChannelFamily,
			Topic:  chat.ID.String(),
		}, mvplive.Envelope{
			EventID:  eventID,
			DedupKey: msg.ID.String(),
		}, func(stream *hotwired.Stream) {
			stream.Replace(msg.HTMLElementID(), string(content))
		})

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
