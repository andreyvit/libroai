package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/edb"
)

func (app *App) showChatHome(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	wls := edb.All(edb.TableScan[m.Waitlister](rc, edb.FullScan()))
	users := edb.All(edb.TableScan[m.User](rc, edb.FullScan()))

	users = append(users, users...)
	users = append(users, users...)
	users = append(users, users...)

	return &mvp.ViewData{
		View:         "superadmin/home",
		Title:        "Chat",
		SemanticPath: "chat",
		Data: struct {
			Waitlisters []*m.Waitlister
			Users       []*m.User
		}{
			Waitlisters: wls,
			Users:       users,
		},
	}, nil
}
