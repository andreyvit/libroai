package main

import "github.com/andreyvit/buddyd/mvp"

func (app *App) registerRoutes(b *mvp.RouteBuilder) {
	b.UseIn("authenticate", app.AuthenticateRequestMiddleware)
	b.Static("/static")
	b.Route("landing.home", "GET /", app.showLandingHome)
	b.Route("landing.signup", "POST /start", app.handleLandingSignup)
	b.Route("landing.waitlist", "GET /waitlist/", app.showWaitlist)
	b.Route("test", "GET /test/", app.showTestPage)
	b.Route("signin", "GET /signin/", app.showSignIn)
	b.Route("signin.process", "POST /signin/", app.handleSignIn, mvp.RateLimitPresetSpam)
	b.Route("signout", "POST /signout/", app.handleSignOut)

	b.Route("accountpicker", "GET /pick-account/", app.showPickAccountForm)

	b.Route("chat.home", "GET /chat/", app.showChatHome)

	b.Group("/lib", func(b *mvp.RouteBuilder) {
		b.Route("lib.home", "GET /", app.showLibraryHome)
	})

	b.Group("/mod", func(b *mvp.RouteBuilder) {
		b.Route("mod.home", "GET /", app.showModerationHome)
	})

	b.Group("/admin", func(b *mvp.RouteBuilder) {
		b.Route("admin.home", "GET /", app.showAdminHome)
	})

	b.Group("/superadmin", func(b *mvp.RouteBuilder) {
		b.UseIn("authorize", fullRC.WrapAE(requireSuperadmin))
		b.Route("superadmin.home", "GET /", app.showSuperadminHome)
		// b.Route("superadmin.superadmins.save", "POST /superadmins/", app.saveSuperadmin)

		b.Group("/db", func(b *mvp.RouteBuilder) {
			b.Route("db.tables", "GET /", app.listSuperadminTables)
			b.Route("db.table.list", "GET /:table/", app.listSuperadminTableRows)
			b.Route("db.table.show", "GET /:table/rows/:key", app.handleSuperadminTableRowForm)
			b.Route("db.table.save", "POST /:table/rows/:key", app.handleSuperadminTableRowForm)

			// b.Route("superadmin.db.dump.simple", "GET /dump.txt", app.showDBRows)
			// b.Route("superadmin.db.dump.full", "GET /dump2.txt", app.showDBDump)
			// b.Route("superadmin.db.dump.stats", "GET /stats.txt", app.showDBStats)

			// g.bg.Handle("GET", "/backup", app.downloadDBBackup)
		})
	})
}
