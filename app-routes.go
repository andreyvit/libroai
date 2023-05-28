package main

import "github.com/andreyvit/buddyd/mvp"

func (app *App) registerRoutes(b *mvp.RouteBuilder) {
	b.Static("/static")

	b.UseIn("authenticate", app.AuthenticateRequestMiddleware)
	b.Use(fullRC.WrapAE(app.initAccountMiddleware))

	b.Route("landing.home", "GET /", app.showLandingHome)
	b.Route("landing.signup", "POST /start", app.handleLandingSignup)
	b.Route("landing.waitlist", "GET /waitlist/", app.showWaitlist)
	b.Route("test", "GET /test/", app.showTestPage)
	b.Route("signin", "GET /signin/", app.showSignIn)
	b.Route("signin.process", "POST /signin/", app.handleSignIn, mvp.RateLimitPresetSpam)
	b.Route("signout", "POST /signout/", app.handleSignOut)

	b.Route("accountpicker", "GET /pick-account/", app.showPickAccountForm)

	b.Group("/chat", func(b *mvp.RouteBuilder) {
		b.UseIn("authorize", fullRC.WrapAE(requireLoggedIn))
		b.Use(fullRC.WrapAE(loadUserChatListMiddleware))

		b.Route("chat.home", "GET /", app.showNewChat)
		b.Route("chat.view", "GET /c/:chat", app.showChat)
		b.Route("chat.messages.send", "POST /c/:chat/send", app.sendChatMessage)
		b.Route("chat.messages.vote", "POST /c/:chat/vote", app.voteChatResponse)

		b.Route("chat.sse", "GET /c/:chat/events/", app.handleChatEventStream).UseIn("authorize", nil)
	})

	b.Group("/lib", func(b *mvp.RouteBuilder) {
		b.UseIn("authorize", fullRC.WrapAE(requireAdmin))
		b.Use(fullRC.WrapAE(loadAccountLibraryMiddleware))

		b.Route("lib.home", "GET /", app.showLibraryRootFolder)
		b.Route("lib.folder", "GET /folders/:folder/", app.showLibraryFolder)
		b.Route("lib.item", "GET /items/:item/", app.showLibraryItem)
	})

	b.Group("/mod", func(b *mvp.RouteBuilder) {
		b.UseIn("authorize", fullRC.WrapAE(requireAdmin))

		b.Route("mod.home", "GET /", app.showModerationHome)
	})

	b.Group("/admin", func(b *mvp.RouteBuilder) {
		b.UseIn("authorize", fullRC.WrapAE(requireAdmin))

		b.Route("admin.home", "GET /", app.showAdminHome)
		b.Route("admin.whitelist", "GET /whitelist/", app.handleAdminWhitelist)
		b.Route("admin.whitelist.save", "POST /whitelist/", app.handleAdminWhitelist)
	})

	b.Group("/superadmin", func(b *mvp.RouteBuilder) {
		b.UseIn("authorize", fullRC.WrapAE(requireSuperadmin))

		b.Route("superadmin.home", "GET /", app.showSuperadminHome)
		// b.Route("superadmin.superadmins.save", "POST /superadmins/", app.saveSuperadmin)

		b.Group("/maintenance", func(b *mvp.RouteBuilder) {
			b.Route("superadmin.maintenance", "GET /", app.listSuperadminProcedures)
			b.Route("superadmin.maintenance.run", "POST /:procedure/", app.runSuperadminProcedure)
		})

		b.Group("/db", func(b *mvp.RouteBuilder) {
			b.Route("db.tables", "GET /", app.listSuperadminTables)
			b.Route("db.table.list", "GET /:table/", app.listSuperadminTableRows)
			b.Route("db.table.show", "GET /:table/rows/:key", app.handleSuperadminTableRowForm)
			b.Route("db.table.save", "POST /:table/rows/:key", app.handleSuperadminTableRowForm)

			// b.Route("superadmin.db.dump.simple", "GET /dump.txt", app.showDBRows)
			b.Route("superadmin.db.dump.full", "GET /fulldump.txt", app.showDBDump)
			// b.Route("superadmin.db.dump.stats", "GET /stats.txt", app.showDBStats)

			// g.bg.Handle("GET", "/backup", app.downloadDBBackup)
		})
	})
}
