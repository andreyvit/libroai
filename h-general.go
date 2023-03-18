package main

func (app *App) showTestPage(rc *RC, in *struct{}) (*ViewData, error) {
	return &ViewData{
		View:  "test",
		Title: "Test Page",
		Data:  struct{}{},
	}, nil
}
