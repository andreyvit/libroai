package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/edb"
)

func (app *App) showLibraryRootFolder(rc *RC, in *struct{}) (*mvp.ViewData, error) {
	return app.doShowLibraryFolder(rc, rc.Library.RootFolderID)
}

func (app *App) showLibraryFolder(rc *RC, in *struct {
	FolderID m.FolderID `form:"folder,path" json:"-"`
}) (*mvp.ViewData, error) {
	return app.doShowLibraryFolder(rc, in.FolderID)
}

func (app *App) doShowLibraryFolder(rc *RC, folderID m.FolderID) (*mvp.ViewData, error) {
	folder := rc.Library.Folder(folderID)
	items := edb.All(edb.ExactIndexScan[m.Item](rc, ItemsByFolder, folderID))
	vm := &m.FolderWithItemsVM{
		Folder: folder,
		Items:  items,
	}
	return &mvp.ViewData{
		View:         "lib/folder",
		Title:        "Library",
		SemanticPath: folder.SemanticPath(),
		Data: struct {
			Folder *m.FolderWithItemsVM
		}{
			Folder: vm,
		},
	}, nil
}
