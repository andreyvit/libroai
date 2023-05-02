package mvp

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"sync/atomic"

	"github.com/andreyvit/buddyd/internal/postmark"
	"github.com/andreyvit/edb"
	"github.com/uptrace/bunrouter"
)

type Module = func(app *App, hooks *Hooks)

type AppOptions struct {
	Context context.Context
	Logf    func(format string, v ...interface{})
}

type AppBehaviors struct {
	IsTesting           bool
	ServeAssetsFromDisk bool
	CrashOnPanic        bool
	PrettyJSON          bool
}

type App struct {
	ValueSet
	Settings *Settings
	Hooks    Hooks
	BaseURL  *url.URL

	stopApp func()
	logf    func(format string, args ...any)

	routesByName map[string]*Route
	domainRouter *DomainRouter
	siteRouters  map[*Site]*bunrouter.Router

	staticFS     fs.FS
	viewsFS      fs.FS
	templates    *template.Template
	templatesDev atomic.Value

	db *edb.DB

	postmrk *postmark.Caller
}

func (app *App) Initialize(settings *Settings, opt AppOptions) {
	if opt.Logf == nil {
		opt.Logf = log.Printf
	}
	if opt.Context == nil {
		opt.Context = context.Background()
	}
	if settings.Env == "" {
		panic("env not set")
	}

	ctx, stopApp := context.WithCancel(opt.Context)
	_ = ctx

	app.ValueSet = newValueSet()
	app.Settings = settings
	app.routesByName = make(map[string]*Route)
	app.logf = opt.Logf
	app.stopApp = stopApp

	if app.BaseURL == nil && settings.BaseURL != "" {
		app.BaseURL = must(url.Parse(settings.BaseURL))
	}

	initAppDB(app, &opt)
	initViews(app, &opt)
	initRouting(app)
	runHooksFwd1(app.Hooks.initApp, app)
}

func (app *App) Close() {
	app.stopApp()
	runHooksRev1(app.Hooks.closeApp, app)
	closeAppDB(app)
}
