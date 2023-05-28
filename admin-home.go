package main

import (
	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"

	m "github.com/andreyvit/buddyd/model"
)

func (app *App) showAdminHome(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	wls := edb.All(edb.TableScan[m.Waitlister](rc, edb.FullScan()))
	users := edb.All(edb.TableScan[m.User](rc, edb.FullScan()))

	return &mvp.ViewData{
		View:         "superadmin/home",
		Title:        "Admin",
		SemanticPath: "admin",
		Data: struct {
			Waitlisters []*m.Waitlister
			Users       []*m.User
		}{
			Waitlisters: wls,
			Users:       users,
		},
	}, nil
}
