package main

func (app *App) registerWebRoutes(r router) {
	r.Add("test", "GET /test/", app.showTestPage)
	// r.Add("home", "GET /chats/:chat", app.showChat)
	// r.Add("home", "POST /send", app.sendChatMessage)
	// r.Add("home", "POST /vote", app.voteChatResponse)
}
