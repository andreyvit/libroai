package main

import (
	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/expandable"
	"github.com/andreyvit/mvp/mvpjobs"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

var (
	appSchema    = expandable.NewSchema("app")
	fullSettings = expandable.Derive[Settings](appSchema, mvp.BaseSettings)
	fullApp      = expandable.Derive[App](appSchema, mvp.BaseApp).WithNew(newApp)
	fullRC       = expandable.Derive[RC](appSchema, mvp.BaseRC)

	dbSchema  = &edb.Schema{}
	jobSchema = &mvpjobs.Schema{}
)

var AppModule = &mvp.Module{
	Name:       "libroai",
	SetupHooks: fullApp.Wrap(setupHooks),
	LoadSecrets: func(*mvp.Settings, mvp.Secrets) {
	},
	DBSchema:  dbSchema,
	JobSchema: jobSchema,
	Types:     map[mvpm.Type][]string{},
}

func setupHooks(app *App) {
	app.Hooks.InitApp(fullApp.Wrap(initApp))
	app.Hooks.InitDB(expandable.Wrap2(initDB, fullApp, fullRC))
	app.Hooks.MakeRowKey(expandable.Wrap21A(makeRowKey, fullApp))
	app.Hooks.ResetAuth(expandable.Wrap2(resetAuth, fullApp, fullRC))
	app.Hooks.PostAuth(expandable.Wrap2E(loadSessionAndUser, fullApp, fullRC))
	app.Hooks.SiteRoutes(mvp.DefaultSite, app.registerRoutes)
	app.Hooks.Helpers(app.registerViewHelpers)
}
