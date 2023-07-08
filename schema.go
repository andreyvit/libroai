package main

import (
	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flake"

	m "github.com/andreyvit/buddyd/model"
)

var (
	UserSignInAttempts = edb.AddTable[m.UserSignInAttempt](dbSchema, "UserSignInAttempt", 1, func(row *m.UserSignInAttempt, ib *edb.IndexBuilder) {
	}, func(tx *edb.Tx, row *m.UserSignInAttempt, oldVer uint64) {
	}, []*edb.Index{})

	Accounts = edb.AddTable(dbSchema, "accounts", 1, func(row *m.Account, ib *edb.IndexBuilder) {
	}, func(tx *edb.Tx, row *m.Account, oldVer uint64) {
	}, []*edb.Index{})

	Users = edb.AddTable(dbSchema, "users", 1, func(row *m.User, ib *edb.IndexBuilder) {
		for _, m := range row.Memberships {
			ib.Add(UsersByAccount, m.AccountID)
		}
		ib.Add(UsersByEmail, row.EmailNorm)
	}, func(tx *edb.Tx, row *m.User, oldVer uint64) {
	}, []*edb.Index{
		UsersByAccount,
		UsersByEmail,
	})
	UsersByAccount = edb.AddIndex[flake.ID]("by_account")
	UsersByEmail   = edb.AddIndex[string]("by_email")

	// Superadmins = edb.AddTable(dbSchema, "superadmins", 1, func(row *m.Superadmin, ib *edb.IndexBuilder) {
	// 	ib.Add(SuperadminsByEmail, row.EmailNorm)
	// }, func(tx *edb.Tx, row *m.Superadmin, oldVer uint64) {
	// }, []*edb.Index{
	// 	SuperadminsByEmail,
	// })
	// SuperadminsByEmail = edb.AddIndex[string]("by_email")

	Waitlisters = edb.AddTable(dbSchema, "waitlisters", 1, func(row *m.Waitlister, ib *edb.IndexBuilder) {
		ib.Add(WaitlistersByEmail, row.EmailNorm)
	}, func(tx *edb.Tx, row *m.Waitlister, oldVer uint64) {
	}, []*edb.Index{
		WaitlistersByEmail,
	})
	WaitlistersByEmail = edb.AddIndex[string]("by_email")

	Sessions = edb.AddTable(dbSchema, "sessions", 1, func(row *m.Session, ib *edb.IndexBuilder) {
		ib.Add(SessionsByActor, row.Actor.ID)
		if row.AccountID != 0 {
			ib.Add(SessionsByAccount, row.AccountID)
		}
	}, func(tx *edb.Tx, row *m.Session, oldVer uint64) {
	}, []*edb.Index{
		SessionsByActor,
		SessionsByAccount,
	})
	SessionsByActor   = edb.AddIndex[flake.ID]("by_actor")
	SessionsByAccount = edb.AddIndex[flake.ID]("by_account")

	Chats = edb.AddTable(dbSchema, "chats", 1, func(row *m.Chat, ib *edb.IndexBuilder) {
		ib.Add(ChatsByAccount, row.AccountID)
		ib.Add(ChatsByUser, row.UserID)
	}, func(tx *edb.Tx, row *m.Chat, oldVer uint64) {
	}, []*edb.Index{
		ChatsByUser,
		ChatsByAccount,
	})
	ChatsByAccount = edb.AddIndex[m.AccountID]("by_account")
	ChatsByUser    = edb.AddIndex[m.UserID]("by_user")

	ChatContent = edb.AddTable(dbSchema, "chat_content_02", 1, func(row *m.ChatContent, ib *edb.IndexBuilder) {
	}, func(tx *edb.Tx, row *m.ChatContent, oldVer uint64) {
	}, []*edb.Index{},
		edb.SuppressContentWhenLogging)

	Folders = edb.AddTable(dbSchema, "folders", 1, func(row *m.Folder, ib *edb.IndexBuilder) {
		ib.Add(FoldersByAccountParent, m.AccountObject(row.AccountID, row.ParentID))
	}, func(tx *edb.Tx, row *m.Folder, oldVer uint64) {
	}, []*edb.Index{
		FoldersByAccountParent,
	})
	FoldersByAccountParent = edb.AddIndex[m.AccountObjectKey]("by_account_parent")

	Items = edb.AddTable(dbSchema, "items", 1, func(row *m.Item, ib *edb.IndexBuilder) {
		ib.Add(ItemsByAccount, row.AccountID)
		ib.Add(ItemsByFolder, row.FolderID)
	}, func(tx *edb.Tx, row *m.Item, oldVer uint64) {
	}, []*edb.Index{
		ItemsByAccount,
		ItemsByFolder,
	})
	ItemsByAccount = edb.AddIndex[m.AccountID]("by_account")
	ItemsByFolder  = edb.AddIndex[m.FolderID]("by_folder")

	Content = edb.AddTable(dbSchema, "content", 1, func(row *m.Content, ib *edb.IndexBuilder) {
		ib.Add(ContentByAccount, row.AccountID)
		ib.Add(ContentByIRO, m.ContentIROKey{
			ItemID:  row.ItemID,
			Role:    row.Role,
			Ordinal: row.Ordinal,
		})
	}, func(tx *edb.Tx, row *m.Content, oldVer uint64) {
	}, []*edb.Index{
		ContentByAccount,
		ContentByIRO,
	})
	ContentByAccount = edb.AddIndex[m.AccountID]("by_account")
	ContentByIRO     = edb.AddIndex[m.ContentIROKey]("by_iro")

	Embeddings = edb.AddTable(dbSchema, "embeddings", 1, func(row *m.ContentEmbedding, ib *edb.IndexBuilder) {
		ib.Add(EmbeddingsByAccountType, m.ContentEmbeddingAccountTypeKey{AccountID: row.AccountID, Type: row.Type})
		ib.Add(EmbeddingsByItem, row.ItemID)
	}, func(tx *edb.Tx, row *m.ContentEmbedding, oldVer uint64) {
	}, []*edb.Index{
		EmbeddingsByAccountType,
		EmbeddingsByItem,
	})
	EmbeddingsByAccountType = edb.AddIndex[m.ContentEmbeddingAccountTypeKey]("by_account_type")
	EmbeddingsByItem        = edb.AddIndex[m.ItemID]("by_item")
)
