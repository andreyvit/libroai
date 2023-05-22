package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
)

func loadAccountLibrary(rc *RC, accountID m.AccountID) *m.AccountLibrary {
	folders := edb.All(edb.IndexScan[m.Folder](rc, FoldersByAccountParent, edb.ExactScan(m.AccountObjectKey{AccountID: accountID}).Prefix(1)))

	library := &m.AccountLibrary{
		RootFolderID:  0,
		Folders:       make(map[m.FolderID]*m.Folder, len(folders)),
		FoldersBySlug: make(map[string]*m.Folder, len(folders)),
	}

	for _, fldr := range folders {
		if fldr.ParentID == 0 {
			library.RootFolderID = fldr.ID
		}
		library.Folders[fldr.ID] = fldr
		library.FoldersBySlug[fldr.Slug] = fldr
	}
	return library
}

func loadCurrentAccountLibrary(rc *RC) {
	rc.Library = loadAccountLibrary(rc, rc.AccountID())
}

func loadAccountLibraryMiddleware(rc *RC) (any, error) {
	loadCurrentAccountLibrary(rc)
	return nil, nil
}
