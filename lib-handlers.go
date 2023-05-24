package main

import (
	"sort"

	"github.com/andreyvit/buddyd/internal/httperrors"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/edb"
	"golang.org/x/exp/maps"
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
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	vm := &m.FolderWithItemsVM{
		Folder: folder,
		Items:  items,
	}
	return &mvp.ViewData{
		View:         "lib/folder",
		Title:        folder.Name,
		SemanticPath: folder.SemanticPath(),
		Data: struct {
			Folder *m.FolderWithItemsVM
		}{
			Folder: vm,
		},
	}, nil
}

func (app *App) showLibraryItem(rc *RC, in *struct {
	ItemID m.ItemID `form:"item,path" json:"-"`
}) (*mvp.ViewData, error) {
	item := edb.Get[m.Item](rc, in.ItemID)
	if item == nil {
		return nil, httperrors.Errorf(404, "", "Item not found")
	}

	fldr := rc.Library.Folder(item.FolderID)

	contents := edb.All(edb.PrefixIndexScan[m.Content](rc, ContentByIRO, 1, m.ContentIROKey{ItemID: item.ID}))

	groupsByRole := make(map[m.ContentRole]*m.ContentGroupVM)
	for _, c := range contents {
		group := groupsByRole[c.Role]
		if group == nil {
			group = &m.ContentGroupVM{
				Role: c.Role,
			}
			groupsByRole[c.Role] = group
		}
		group.Contents = append(group.Contents, c)
	}

	groups := maps.Values(groupsByRole)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Role < groups[j].Role
	})

	return &mvp.ViewData{
		View:         "lib/item",
		Title:        item.Name,
		SemanticPath: item.SemanticPath(),
		Data: struct {
			Folder        *m.Folder
			Item          *m.Item
			ContentGroups []*m.ContentGroupVM
		}{
			Folder:        fldr,
			Item:          item,
			ContentGroups: groups,
		},
	}, nil
}
