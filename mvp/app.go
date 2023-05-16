package mvp

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/andreyvit/buddyd/internal/postmark"
	"github.com/andreyvit/buddyd/mvp/flake"
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
	DisableRateLimits   bool
	AllowInsecureHttp   bool
}

type App struct {
	ValueSet
	Configuration *Configuration
	Settings      *Settings
	Hooks         Hooks
	BaseURL       *url.URL

	stopApp func()
	logf    func(format string, args ...any)

	routesByName map[string]*Route
	domainRouter *DomainRouter
	siteRouters  map[*Site]*bunrouter.Router

	staticFS     fs.FS
	viewsFS      fs.FS
	templates    *template.Template
	templatesDev atomic.Value

	db  *edb.DB
	gen *flake.Gen

	postmrk *postmark.Caller

	rateLimiters map[RateLimitPreset]map[RateLimitGranularity]*RateLimiter

	// rateLimiters map[string]
}

func (app *App) Initialize(settings *Settings, opt AppOptions) {
	if opt.Logf == nil {
		opt.Logf = log.Printf
	}
	if opt.Context == nil {
		opt.Context = context.Background()
	}
	if settings.Env == "" {
		panic("settings.Env not set")
	}
	if settings.AppID == "" {
		panic(fmt.Errorf("%s: AppID not configured", settings.Configuration.ConfigFileName))
	}
	if len(settings.JWTIssuers) == 0 {
		settings.JWTIssuers = []string{settings.AppID}
	}

	ctx, stopApp := context.WithCancel(opt.Context)
	_ = ctx

	app.ValueSet = newValueSet()
	app.Configuration = settings.Configuration
	app.Settings = settings
	app.routesByName = make(map[string]*Route)
	app.logf = opt.Logf
	app.stopApp = stopApp

	if app.BaseURL == nil && settings.BaseURL != "" {
		app.BaseURL = must(url.Parse(settings.BaseURL))
	}

	initAppDB(app, &opt)
	initViews(app, &opt)

	app.postmrk = &postmark.Caller{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Credentials: app.Settings.Postmark,
	}

	initRateLimiting(app)
	initRouting(app)
	runHooksFwd1(app.Hooks.initApp, app)

	{
		rc := NewRC(ctx, app, "init")
		err := app.InTx(rc, true, func() error {
			runHooksFwd2(app.Hooks.initDB, app, rc)
			return nil
		})
		if err != nil {
			panic(fmt.Errorf("db init failed: %v", err))
		}
	}
}

func (app *App) Close() {
	app.stopApp()
	runHooksRev1(app.Hooks.closeApp, app)
	closeAppDB(app)
}
