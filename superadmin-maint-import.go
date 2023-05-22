package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreyvit/buddyd/internal/flogger"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp/forms"
	"github.com/andreyvit/edb"
)

func (app *App) importProcedure() *Procedure {
	type legacyMemoryEntry struct {
		Text          string    `json:"text"`
		TextTokens    int       `json:"tokens"`
		TextEmbedding []float64 `json:"embedding"`
	}

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
			flogger.Log(rc, "Root: %s", in.Path)

			loadCurrentAccountLibrary(rc)

			fldr := rc.Library.FolderBySlug("imported")
			if fldr == nil {
				fldr = &m.Folder{
					ID:        app.NewID(),
					AccountID: rc.AccountID(),
					Name:      "Imported",
					Slug:      "imported",
					ParentID:  rc.Library.RootFolderID,
				}
				edb.Put(rc, fldr)
				parent := rc.Library.RootFolder()
				parent.ChildenIDs = append(parent.ChildenIDs, fldr.ID)
				edb.Put(rc, parent)
			}

			items := edb.All(edb.ExactIndexScan[m.Item](rc, ItemsByFolder, fldr.ID))
			existingItemsByFileName := make(map[string]*m.Item)
			for _, item := range items {
				if item.FileName != "" {
					existingItemsByFileName[item.FileName] = item
				}
			}

			memEmbPath := filepath.Join(in.Path, "memory-embeddings")
			flogger.Log(rc, "Importing embeddings from %s", memEmbPath)
			files := collectFiles(memEmbPath, ".json")

			for _, fn := range files {
				base := filepath.Base(fn)
				base = strings.TrimSuffix(base, ".json")

				flogger.Log(rc, "Processing %s", base)

				item := existingItemsByFileName[base]
				delete(existingItemsByFileName, base)
				if item == nil {
					item = &m.Item{
						ID:        app.NewID(),
						AccountID: rc.AccountID(),
						FolderID:  fldr.ID,
						Name:      base,
						FileName:  base,
					}
					edb.Put(rc, item)
				}
			}

			return nil
		},
	}
}

func collectFiles(root string, suffix string) []string {
	var result []string
	ensure(fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		result = append(result, path)
		return nil
	}))
	return result
}
