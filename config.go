package main

import "github.com/andreyvit/openai"

const (
	// DefaultModel              = openai.ModelChatGPT4
	DefaultModel              = openai.ModelChatGPT35Turbo
	MinModel                  = openai.ModelChatGPT35Turbo
	EmbeddingModel            = openai.ModelEmbeddingAda002
	MaxMsgTokenCount          = 768
	MaxSystemPromptTokenCount = 1024
	MaxResponseTokenCount     = 512
)
