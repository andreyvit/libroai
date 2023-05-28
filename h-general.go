package main

import "github.com/andreyvit/mvp"

func (app *App) showTestPage(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	return &mvp.ViewData{
		View:  "test",
		Title: "Test Page",
		Data:  struct{}{},
	}, nil
}
