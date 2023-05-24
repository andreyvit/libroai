package m

import (
	"fmt"
	"sort"
)

const DistanceEps = 1e-6

type AccountEmbeddings struct {
	Embeddings []*ContentEmbedding
}

func (embs *AccountEmbeddings) Select(questionEmbedding []float64, maxCount int, maxDistance float64) EntriesAndDistances {
	entries := make([]*ContentEmbedding, len(embs.Embeddings))
	copy(entries, embs.Embeddings) // we'll be sorting this so make a copy

	distances := make([]float64, len(entries))
	for i, e := range entries {
		distances[i] = CosineDistance(questionEmbedding, e.Embedding)
	}

	return EntriesAndDistances{entries, distances}.SelectTop(maxCount, maxDistance)
}

func (ed *EntriesAndDistances) Find(contentID ContentID) int {
	for i, e := range ed.Entries {
		if e.ContentID == contentID {
			return i
		}
	}
	return -1
}

func (ed *EntriesAndDistances) AppendAll(more EntriesAndDistances) {
	orig := *ed
	for i, e := range more.Entries {
		d := more.Distances[i]
		if i := orig.Find(e.ContentID); i >= 0 {
			if d < ed.Distances[i] {
				ed.Distances[i] = d
			}
		} else {
			ed.Entries = append(ed.Entries, e)
			ed.Distances = append(ed.Distances, d)
		}
	}
}

func (ed EntriesAndDistances) SelectTop(maxCount int, maxDistance float64) EntriesAndDistances {
	n := len(ed.Entries)
	sort.Sort(EntriesAndDistances{ed.Entries, ed.Distances})

	cutoff := maxCount
	if cutoff > n {
		cutoff = n
	}
	maxDistance -= DistanceEps
	for cutoff > 0 && ed.Distances[cutoff-1] > maxDistance {
		cutoff--
	}

	return EntriesAndDistances{ed.Entries[:cutoff], ed.Distances[:cutoff]}
}

type EntriesAndDistances struct {
	Entries   []*ContentEmbedding
	Distances []float64
}

func (a EntriesAndDistances) DistancesByContentID() map[ContentID]float64 {
	result := make(map[ContentID]float64, len(a.Entries))
	for i, e := range a.Entries {
		result[e.ContentID] = a.Distances[i]
	}
	return result
}

func (a EntriesAndDistances) Len() int { return len(a.Distances) }
func (a EntriesAndDistances) Swap(i, j int) {
	a.Distances[i], a.Distances[j] = a.Distances[j], a.Distances[i]
	a.Entries[i], a.Entries[j] = a.Entries[j], a.Entries[i]
}
func (a EntriesAndDistances) Less(i, j int) bool { return a.Distances[i] > a.Distances[j] }

// CosineDistance returns a measure of similarity of the given embeddings (1 equal ... 0 unrelated ... -1 opposite).
// This is basically a cosine of the angle between the embedding vectors.
// Assumes the embedding vectors are coming from OpenAI and are normalized,
// so uses dot product to compute the distance.
func CosineDistance(a, b []float64) float64 {
	var sum float64
	if len(a) != len(b) {
		panic(fmt.Errorf("dot product requires equal-length vectors: %d vs %d", len(a), len(b)))
	}
	for i, v := range a {
		sum += v * b[i]
	}
	return sum
}
