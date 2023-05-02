package mvp

import (
	"html/template"
)

type Hooks struct {
	initApp      []func(app *App)
	closeApp     []func(app *App)
	initRC       []func(app *App, rc *RC)
	closeRC      []func(app *App, rc *RC)
	helpers      []func(m template.FuncMap)
	middleware   []func(r Router)
	domainRoutes []func(app *App, b *DomainRouter)
	siteRoutes   map[*Site][]func(b *RouteBuilder)
}

func (h *Hooks) InitApp(f func(app *App)) {
	h.initApp = append(h.initApp, f)
}

func (h *Hooks) CloseApp(f func(app *App)) {
	h.closeApp = append(h.closeApp, f)
}

func (h *Hooks) InitRC(f func(app *App, rc *RC)) {
	h.initRC = append(h.initRC, f)
}

func (h *Hooks) CloseRC(f func(app *App, rc *RC)) {
	h.closeRC = append(h.closeRC, f)
}

func (h *Hooks) Helpers(f func(m template.FuncMap)) {
	h.helpers = append(h.helpers, f)
}

func (h *Hooks) Middleware(f func(r Router)) {
	h.middleware = append(h.middleware, f)
}

func (h *Hooks) DomainRoutes(f func(app *App, b *DomainRouter)) {
	h.domainRoutes = append(h.domainRoutes, f)
}

func (h *Hooks) SiteRoutes(site *Site, f func(b *RouteBuilder)) {
	if h.siteRoutes == nil {
		h.siteRoutes = make(map[*Site][]func(b *RouteBuilder))
	}
	h.siteRoutes[site] = append(h.siteRoutes[site], f)
}

func runHooksFwd1[T1 any](hooks []func(a1 T1), a1 T1) {
	for _, f := range hooks {
		f(a1)
	}
}

func runHooksRev1[T1 any](hooks []func(a1 T1), a1 T1) {
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i](a1)
	}
}

func runHooksFwd2[T1, T2 any](hooks []func(a1 T1, a2 T2), a1 T1, a2 T2) {
	for _, f := range hooks {
		f(a1, a2)
	}
}

func runHooksRev2[T1, T2 any](hooks []func(a1 T1, a2 T2), a1 T1, a2 T2) {
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i](a1, a2)
	}
}
