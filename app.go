package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/andreyvit/buddyd/internal/accesstokens"
)

type AppOptions struct {
	Logf func(format string, v ...interface{})
}

type App struct {
	BaseURL *url.URL
	Logf    func(format string, v ...interface{})

	Store *Store

	users atomic.Value

	routesByName map[string]*routeInfo
	siteData     *SiteData
	templates    *template.Template

	webAdminTokens accesstokens.Configuration

	// memory       *vdemir.Memory
	httpClient *http.Client
}

func setupApp(dataDir string, opt AppOptions) *App {
	ensure(os.MkdirAll(dataDir, 0755))

	if opt.Logf == nil {
		opt.Logf = log.Printf
	}

	app := &App{
		BaseURL: must(url.Parse(settings.BaseURL)),
		Logf:    opt.Logf,

		Store: setupStore(dataDir, 1),

		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}

	app.routesByName = make(map[string]*routeInfo)
	app.siteData = &SiteData{
		AppName: settings.AppName,
	}
	app.templates = must(app.loadTemplates())

	return app
}

func (app *App) Close() {
}
