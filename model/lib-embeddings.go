package m

import "github.com/andreyvit/openai"

type Embedding = []float64

type ContentEmbedding struct {
	ContentEmbeddingKey `msgpack:"-"`
	AccountID           AccountID `msgpack:"a"`
	ItemID              ItemID    `msgpack:"i"`
	TokenCountGPT35     int       `msgpack:"t3"`
	Embedding           `msgpack:"e"`
}

func (emb *ContentEmbedding) UpdateTokenCount(c *Content) {
	emb.TokenCountGPT35 = openai.TokenCount(c.Text, openai.ModelChatGPT35Turbo)
}

func (emb *ContentEmbedding) TokenCount(model string) int {
	return emb.TokenCountGPT35
}

type ContentEmbeddingKey struct {
	ContentID ContentID
	Type      EmbeddingType
}

type ContentEmbeddingAccountTypeKey struct {
	AccountID AccountID
	Type      EmbeddingType
}
