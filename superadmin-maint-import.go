package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/andreyvit/buddyd/internal/flogger"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp/forms"
	"github.com/andreyvit/edb"
)

func (app *App) importProcedure() *Procedure {
	in := &struct {
		Path string
	}{}

	return &Procedure{
		Slug:  "import-legacy",
		Title: "Import Legacy Library",
		Form: &forms.Form{
			Group: forms.Group{
				Styles: []*forms.Style{
					adminFormStyle,
					verticalFormStyle,
				},
				Children: []forms.Child{
					&forms.Item{
						Name:  "path",
						Label: "Local Path",
						Child: &forms.InputText{
							Binding:     forms.Var(&in.Path),
							Placeholder: "/Users/foo/bar/boz",
						},
					},
				},
			},
		},
		Handler: func(rc *RC) error {
			in.Path = "/Users/andreyvit/Developer/lifehack/virtual-demir/_library"
			flogger.Log(rc, "Root: %s", in.Path)

			loadCurrentAccountLibrary(rc)

			importedFolder := ensureFolderBySlug(rc, "imported", "Imported", rc.Library.RootFolderID)

			items := edb.All(edb.ExactIndexScan[m.Item](rc, ItemsByAccount, rc.AccountID()))
			existingImportedItemsBySourceName := make(map[string]*m.Item)
			for _, item := range items {
				if item.ImportSourceName != "" {
					existingImportedItemsBySourceName[item.ImportSourceName] = item
				}
			}

			memEmbPath := filepath.Join(in.Path, "memory-embeddings")
			flogger.Log(rc, "Importing embeddings from %s", memEmbPath)
			files := collectFiles(memEmbPath, ".json")

			iis := make(map[string]*importableItem)

			for _, fn := range files {
				base := filepath.Base(fn)
				base = strings.TrimSuffix(base, ".json")

				var sourceName, subfolderName, subfolderSlug, displayName string
				if m := faqlibBaseRe.FindStringSubmatch(base); m != nil {
					sourceName = m[1]
					subfolderSlug = m[2]
					subfolderName = fmt.Sprintf("FAQ %s", m[3])
				} else if m := videoRe.FindStringSubmatch(base); m != nil {
					sourceName = m[1]
					subfolderName = fmt.Sprintf("BC%s", m[2])
					subfolderSlug = fmt.Sprintf("bc%s", m[2])
					displayName = fmt.Sprintf("BC%s %s (%s)", m[2], m[3], m[4])
				} else if m := recapsRe.FindStringSubmatch(base); m != nil {
					sourceName = base
					subfolderName = fmt.Sprintf("%s Recaps", m[1])
					subfolderSlug = fmt.Sprintf("%s-recaps", m[1])
					displayName = fmt.Sprintf("%s Recaps %s", m[1], m[2])
				} else if m := importableBaseRe.FindStringSubmatch(base); m != nil {
					sourceName = m[1]
				} else {
					flogger.Log(rc, "WARNING: unmatched: %s", base)
					continue
				}
				if displayName == "" {
					displayName = strings.ReplaceAll(sourceName, "_", " ")
				}

				ii := iis[sourceName]
				if ii == nil {
					ii = &importableItem{
						DisplayName:   displayName,
						SourceName:    sourceName,
						SubfolderName: subfolderName,
						SubfolderSlug: subfolderSlug,
					}
					iis[sourceName] = ii
				}
				ii.Files = append(ii.Files, &importableFile{
					Base: base,
					Path: fn,
				})
			}

			for _, ii := range iis {
				sort.Slice(ii.Files, func(i, j int) bool {
					return ii.Files[i].Base < ii.Files[j].Base
				})
			}

			// flogger.Log(rc, "Importable Items: %s", must(json.MarshalIndent(iis, "", "    ")))

			for _, ii := range iis {
				var subfolder *m.Folder
				if ii.SubfolderSlug == "" {
					subfolder = importedFolder
				} else {
					subfolder = ensureFolderBySlug(rc, ii.SubfolderSlug, ii.SubfolderName, importedFolder.ID)
				}

				item := existingImportedItemsBySourceName[ii.SourceName]
				delete(existingImportedItemsBySourceName, ii.SourceName)
				if item == nil {
					item = &m.Item{
						ID:               app.NewID(),
						AccountID:        rc.AccountID(),
						FolderID:         subfolder.ID,
						Name:             ii.DisplayName,
						ImportSourceName: ii.SourceName,
					}
					edb.Put(rc, item)
				}

				deleteContentByItem(rc, item.ID)
				for i, f := range ii.Files {
					flogger.Log(rc, "Parsing %s", f.Base)
					raw := must(os.ReadFile(f.Path))
					var entry importedMemoryEntry
					ensure(json.Unmarshal(raw, &entry))

					c := &m.Content{
						ID:        app.NewID(),
						AccountID: rc.AccountID(),
						ItemID:    item.ID,
						Role:      m.ContentRoleMemory,
						Ordinal:   i,
						Text:      entry.Text,
					}
					edb.Put(rc, c)

					e := &m.ContentEmbedding{
						ContentEmbeddingKey: m.ContentEmbeddingKey{ContentID: c.ID, Type: m.EmbeddingTypeAda02},
						AccountID:           rc.AccountID(),
						ItemID:              item.ID,
						Embedding:           entry.TextEmbedding,
					}
					edb.Put(rc, e)
				}
			}

			for _, item := range existingImportedItemsBySourceName {
				flogger.Log(rc, "Deleting legacy item %v %s", item.ID, item.Name)
				deleteItem(rc, item.ID)
			}

			return nil
		},
	}
}

type importableItem struct {
	DisplayName   string
	SourceName    string
	SubfolderName string
	SubfolderSlug string
	Files         []*importableFile
}

type importableFile struct {
	Base string
	Path string
}

type importedMemoryEntry struct {
	Text          string    `json:"text"`
	TextTokens    int       `json:"tokens"`
	TextEmbedding []float64 `json:"embedding"`
}

var (
	faqlibBaseRe     = regexp.MustCompile(`^((faqlib-([a-z]+))-\d+)-\d+$`)
	recapsRe         = regexp.MustCompile(`^(.*)-recaps-(\d+)$`)
	importableBaseRe = regexp.MustCompile(`^(.*?)-(Q\d+-v\d+-\d+|C\d+-\d+|\d+-\d+|\d+)$`)
	videoRe          = regexp.MustCompile(`^(bc(\d+)-(\d{4}-\d{2}-\d{2})-(\w{3}))-Q\d+-v\d+-\d+$`)
)

func collectFiles(root string, suffix string) []string {
	var result []string
	ensure(fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		result = append(result, filepath.Join(root, path))
		return nil
	}))
	return result
}
