package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type AppOptions struct {
	Logf func(format string, v ...interface{})
}

type App struct {
	BaseURL *url.URL
	Logf    func(format string, v ...interface{})

	Store *Store

	// State        *botstate.State
	siteData   *SiteData
	webHandler http.Handler
	templates  *template.Template

	// accessTokens forevertokens.Configuration
	// memory       *vdemir.Memory
	httpClient *http.Client
}

func setupApp(dataDir string, opt AppOptions) *App {
	ensure(os.MkdirAll(dataDir, 0755))

	if opt.Logf == nil {
		opt.Logf = log.Printf
	}

	app := &App{
		BaseURL: must(url.Parse(baseURLStr)),
		Logf:    opt.Logf,

		Store: setupStore(dataDir, 1),

		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}

	app.siteData = &SiteData{
		AppName: appName,
	}

	app.webHandler = app.setupSiteRouter(SubsiteWeb)
	app.templates = must(app.loadTemplates())

	return app
}

func (app *App) Close() {
}
