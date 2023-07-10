package main

import "encoding/json"

const (
	chatTitleSystemPrompt = `Make SHORT chat title max 5 words`
	chatTitleFuncName     = "set_chat_title"
)

type ChatTitleFuncResult struct {
	Title string `json:"title"`
}

var (
	chatTitleFunc = json.RawMessage(`{
		"name": "set_chat_title",
		"description": "Set the title of the chat.",
		"parameters": {
			"type": "object",
			"required": [
				"title",
				"subtitle"
			],
			"properties": {
				"title": {
					"type": "string",
					"description": "5 words or less"
				}
			}
		}
	}`)
)
