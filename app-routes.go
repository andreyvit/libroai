package main

import "github.com/andreyvit/buddyd/mvp"

func (app *App) registerRoutes(b *mvp.RouteBuilder) {
	b.Static("/static")
	b.Route("landing.home", "GET /", app.showLandingHome)
	b.Route("landing.signup", "POST /start", app.handleLandingSignup)
	b.Route("landing.waitlist", "GET /waitlist/", app.showWaitlist)
	b.Route("test", "GET /test/", app.showTestPage)
	b.Route("signin", "GET /signin/", app.showSignIn)
	b.Route("signin.process", "POST /signin/", app.handleSignIn, mvp.RateLimitPresetSpam)

	b.Route("accountpicker", "GET /pick-account/", app.showPickAccountForm)

	b.Route("chat.home", "GET /chat/", app.showChatHome)

	b.Group("/admin", func(b *mvp.RouteBuilder) {
		b.Route("admin.home", "GET /", app.showAdminHome)
	})

	b.Group("/superadmin", func(b *mvp.RouteBuilder) {
		b.Route("superadmin.home", "GET /", app.showSuperadminHome)
		// b.Route("superadmin.superadmins.save", "POST /superadmins/", app.saveSuperadmin)
	})
}
