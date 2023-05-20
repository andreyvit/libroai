package main

import "github.com/andreyvit/openai"

const (
	DefaultModel          = openai.ModelChatGPT4
	MinModel              = openai.ModelChatGPT35Turbo
	EmbeddingModel        = "text-embedding-ada-002"
	MaxMsgTokenCount      = 768
	MaxResponseTokenCount = 512
)
