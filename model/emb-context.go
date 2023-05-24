package m

import (
	"strings"

	"github.com/andreyvit/openai"
)

func InsertMessageContent(prefix, suffix, sep string, content []*Content) string {
	var buf strings.Builder
	buf.WriteString(prefix)
	for _, c := range content {
		buf.WriteString(sep)
		buf.WriteString(c.Text)
	}
	if len(suffix) > 0 {
		buf.WriteString(sep)
		buf.WriteString(suffix)
	}
	return buf.String()
}

func PickContext(prefix, suffix, sep string, maxTokens int, availableEntries []*ContentEmbedding, model string) (includedEntries []*ContentEmbedding, usedTokens int) {
	sepTokens := openai.TokenCount(sep, model)
	usedTokens = openai.TokenCount(prefix, model)
	if len(suffix) > 0 {
		usedTokens += sepTokens + openai.TokenCount(suffix, model)
	}
	for _, e := range availableEntries {
		t := sepTokens + e.TokenCount(model)
		if usedTokens+t <= maxTokens {
			includedEntries = append(includedEntries, e)
			usedTokens += t
		}
	}
	return
}
