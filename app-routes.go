package main

func (app *App) registerRoutes(g *routeBuilder) {
	g.static("/static", staticAssetsFS)
	g.route("test", "GET /test/", app.showTestPage)
}
