package main

import (
	"github.com/andreyvit/mvp"
)

func (app *App) showAccountActivity(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	return &mvp.ViewData{
		View:         "mod/activity",
		Title:        "Activity",
		SemanticPath: "mod/activity",
		Data: struct {
		}{},
	}, nil
}
