package main

var (
	usersColl = &collection{
		name:      "chats",
		singleton: true,
		ext:       ".json",
	}
	chatsColl = &collection{
		name: "chats",
		ext:  ".json",
	}

	collections = []*collection{
		usersColl,
		chatsColl,
	}
)
