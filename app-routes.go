package main

import "github.com/andreyvit/buddyd/mvp"

func (app *App) registerRoutes(b *mvp.RouteBuilder) {
	b.Static("/static")
	b.Route("landing.home", "GET /", app.showLandingHome)
	b.Route("landing.signup", "POST /start", app.handleLandingSignup)
	b.Route("landing.waitlist", "GET /waitlist/", app.showWaitlist)
	b.Route("test", "GET /test/", app.showTestPage)
}
