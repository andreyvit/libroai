package main

import (
	"strings"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flogger"

	m "github.com/andreyvit/buddyd/model"
)

const (
	MaxContextEntries          = 15
	MaxContextDistance float64 = 1e6
)

const (
	promptSep = "\n\n---\n\n"
	prompt1   = `You are a helpful assistant bot made by productivity coach Demir Bentley, founder of LifeHack Bootcamp and LifeHack Method. You are responding as Demir. Answer comprehensively. Be concise, but comprehensive. Use the information below. || Help the user concicely.`
)

// [CONTEXT]

// Answer user's question using the above information where possible. Be concise, but comprehensive.
// ---
// [HISTORY]
// ---
// [QUESTION]

type PromptResult struct {
	Prompt            string
	ContextContentIDs []m.ContentID
	ContextDistances  []float64
}

func (app *App) BuildSystemPrompt(rc *RC, prompt string, chat *m.Chat, cc *m.ChatContent) (PromptResult, error) {
	var result PromptResult

	prefix, suffix, _ := strings.Cut(prompt, "||")
	prefix = strings.TrimSpace(prefix)
	suffix = strings.TrimSpace(suffix)

	m1 := cc.FirstUserMessage()
	m2 := cc.LatestUserMessage()

	// flogger.Log(rc, "First message: %v", m1.Text)
	// flogger.Log(rc, "Last message: %v", m2.Text)

	var entries m.EntriesAndDistances
	if m1 != nil {
		ed := rc.Embeddings.Select(m1.EmbeddingAda002, MaxContextEntries, MaxContextDistance)
		// flogger.Log(rc, "First message context: %d", len(ed.Entries))
		entries.AppendAll(ed)
		// flogger.Log(rc, "Total context: %d", len(entries.Entries))
	}
	if m2 != nil && m2 != m1 {
		ed := rc.Embeddings.Select(m2.EmbeddingAda002, MaxContextEntries, MaxContextDistance)
		// flogger.Log(rc, "Second message context: %d", len(ed.Entries))
		entries.AppendAll(ed)
		// flogger.Log(rc, "Total context: %d", len(entries.Entries))
		entries = entries.SelectTop(MaxContextEntries, MaxContextDistance)
		// flogger.Log(rc, "Trimmed context: %d", len(entries.Entries))
	}

	distancesByContentID := entries.DistancesByContentID()

	includedEntries, _ := m.PickContext(prefix, suffix, promptSep, MaxSystemPromptTokenCount, entries.Entries, DefaultModel)
	includedContent := make([]*m.Content, 0, len(includedEntries))

	app.MustRead(rc.BaseRC(), func() {
		for _, e := range includedEntries {
			c := edb.Get[m.Content](rc, e.ContentID)
			if c != nil {
				includedContent = append(includedContent, c)
				result.ContextContentIDs = append(result.ContextContentIDs, c.ID)
				result.ContextDistances = append(result.ContextDistances, distancesByContentID[c.ID])
			} else {
				flogger.Log(rc, "WARNING: entry refers to missing content %v (item %v)", e.ContentID, e.ItemID)
			}
		}
	})

	result.Prompt = m.InsertMessageContent(prefix, suffix, promptSep, includedContent)
	return result, nil
}
