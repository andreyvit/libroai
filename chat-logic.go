package main

import (
	"context"
	"fmt"

	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/openai"
)

const prompt = `You are a helpful assistant bot made by productivity coach Demir Bentley, founder of LifeHack Bootcamp and LifeHack Method. You are responding as Demir. Answer comprehensively. Be concise, but comprehensive.`

func (app *App) ProduceBotMessage(ctx context.Context, chat *m.Chat, content *m.ChatContent) (*m.Message, error) {
	// embedding, usage, err := openai.ComputeEmbedding(context.Background(), msgContent, app.httpClient, app.OpenAICreds)
	// if err != nil {
	// 	return nil, fmt.Errorf("embeddings: %w", err)
	// }
	// round.UserMessageEmbedding = embedding
	// chat.Rounds = append(chat.Rounds, round)
	// chat.Cost += openai.Cost(usage.PromptTokens, usage.CompletionTokens, EmbeddingModel)
	// app.State.SaveChat(chat)

	// contextEntries, contextRanks := app.memory.Select(embedding)

	var history []openai.Msg
	history = append(history, openai.SystemMsg(prompt))
	for _, msg := range content.Messages {
		history = append(history, openai.Msg{
			Role:    msg.Role.OpenAIRole(),
			Content: msg.Text,
		})
	}

	opt := openai.DefaultChatOptions()
	opt.Model = DefaultModel
	opt.MaxTokens = MaxResponseTokenCount
	opt.Temperature = 0.75

	choices, usage, err := openai.Chat(ctx, history, opt, app.httpClient, app.Settings().OpenAICreds)
	if err != nil {
		return nil, err
	}
	chat.Cost += openai.Cost(usage.PromptTokens, usage.CompletionTokens, opt.Model)

	if len(choices) != 1 {
		return nil, fmt.Errorf("len(choices) = %d", len(choices))
	}
	choice := choices[0]

	msg := &m.Message{
		ID:   app.NewID(),
		Role: m.MessageRoleBot,
		Text: choice.Content,
	}
	content.Messages = append(content.Messages, msg)

	return msg, nil
}
