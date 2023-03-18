package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/andreyvit/httpform"
	"github.com/uptrace/bunrouter"
)

type Subsite int

const (
	SubsiteNone Subsite = iota
	SubsiteWeb
)

func (app *App) setupHandler() http.Handler {
	web := app.setupSiteRouter(SubsiteWeb)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch trimPort(r.Host) {
		case app.BaseURL.Host:
			web.ServeHTTP(w, r)
		default:
			http.Error(w, fmt.Sprintf("Invalid domain %q (wanted %s)", trimPort(r.Host), app.BaseURL.Host), http.StatusMisdirectedRequest)
		}
	})
}

func (app *App) setupSiteRouter(subsite Subsite) http.Handler {
	bun := bunrouter.New()

	setupStaticServer(bun, "/static", staticAssetsFS)

	r := router{app, &bun.Group}
	switch subsite {
	case SubsiteWeb:
		app.registerWebRoutes(r)
	}
	return bun
}

func (app *App) callRoute(route *routeInfo, rc *RC, w http.ResponseWriter, req bunrouter.Request) error {
	if !route.Idempotent {
		// TODO: check CSRF
	}

	iv := reflect.New(route.InType)
	err := httpform.Default.DecodeVal(req.Request, req.Params(), iv)
	if err != nil {
		return err
	}

	inputs := make([]reflect.Value, route.FuncVal.Type().NumIn())
	inputs[0] = reflect.ValueOf(rc)
	inputs[1] = iv
	results := route.FuncVal.Call(inputs)
	output := results[0].Interface()
	if errVal := results[1].Interface(); errVal != nil {
		err = errVal.(error)
		if err != nil {
			return err
		}
	}

	return app.returnResponse(rc, output, route, w, req)
}
