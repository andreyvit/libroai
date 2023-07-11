package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
)

func (app *App) initAccount(rc *RC, account *m.Account) {
	rootFolder := edb.Lookup[m.Folder](rc, FoldersByAccountParent, m.AccountObject(account.ID, 0))
	if rootFolder == nil {
		rootFolder = &m.Folder{
			ID:        app.NewID(),
			AccountID: account.ID,
			Name:      "Library",
			Slug:      "library",
			ParentID:  0,
		}
		edb.Put(rc, rootFolder)
	}
}

func (app *App) initAccountMiddleware(rc *RC) (any, error) {
	if rc.DBTx().IsWritable() && rc.Account != nil {
		app.initAccount(rc, rc.Account.Account)
	}
	return nil, nil
}
