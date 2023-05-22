package m

type Embedding = []float64

type ContentEmbedding struct {
	ContentEmbeddingKey `msgpack:"-"`
	AccountID           AccountID `msgpack:"a"`
	Embedding           `msgpack:"e"`
}

type ContentEmbeddingKey struct {
	ContentID ContentID
	Type      EmbeddingType
}

type ContentEmbeddingAccountTypeKey struct {
	AccountID AccountID
	Type      EmbeddingType
}
