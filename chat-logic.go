package main

import (
	"fmt"

	"github.com/andreyvit/buddyd/internal/flogger"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/openai"
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

	pr, err := app.BuildSystemPrompt(rc, prompt1, chat, cc)
	if err != nil {
		return nil, err
	}

	flogger.Log(rc, "Prompt: %s", pr.Prompt)

	var history []openai.Msg
	history = append(history, openai.SystemMsg(pr.Prompt))
	for _, t := range cc.Turns {
		msg := t.LatestVersion()
		history = append(history, openai.Msg{
			Role:    msg.Role.OpenAIRole(),
			Content: msg.Text,
		})
	}

	opt := openai.DefaultChatOptions()
	opt.Model = DefaultModel
	opt.MaxTokens = MaxResponseTokenCount
	opt.Temperature = 0.75

	choices, usage, err := openai.Chat(rc, history, opt, app.httpClient, app.Settings().OpenAICreds)
	if err != nil {
		return nil, err
	}
	chat.Cost += openai.Cost(usage.PromptTokens, usage.CompletionTokens, opt.Model)

	if len(choices) != 1 {
		return nil, fmt.Errorf("len(choices) = %d", len(choices))
	}
	choice := choices[0]

	msg := &m.Message{
		ID:                app.NewID(),
		Role:              m.MessageRoleBot,
		Text:              choice.Content,
		ContextContentIDs: pr.ContextContentIDs,
		ContextDistances:  pr.ContextDistances,
	}
	cc.Turns = append(cc.Turns, &m.Turn{
		Role: m.MessageRoleBot,
		Versions: []*m.Message{
			msg,
		},
	})

	return msg, nil
}
