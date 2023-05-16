package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/edb"
)

func (app *App) showSuperadminHome(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	wls := edb.All(edb.TableScan[m.Waitlister](rc, edb.FullScan()))
	users := edb.All(edb.TableScan[m.User](rc, edb.FullScan()))

	return &mvp.ViewData{
		View:         "superadmin/home",
		Title:        "Superadmin",
		SemanticPath: "superadmin",
		Data: struct {
			Waitlisters []*m.Waitlister
			Users       []*m.User
		}{
			Waitlisters: wls,
			Users:       users,
		},
	}, nil
}
