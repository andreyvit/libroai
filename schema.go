package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp/flake"
	"github.com/andreyvit/edb"
)

var schema = edb.NewSchema(edb.SchemaOpts{})

var (
	UserSignInAttempts = edb.AddTable[m.UserSignInAttempt](schema, "UserSignInAttempt", 1, func(row *m.UserSignInAttempt, ib *edb.IndexBuilder) {
	}, func(tx *edb.Tx, row *m.UserSignInAttempt, oldVer uint64) {
	}, []*edb.Index{})

	Accounts = edb.AddTable(schema, "accounts", 1, func(row *m.Account, ib *edb.IndexBuilder) {
	}, func(tx *edb.Tx, row *m.Account, oldVer uint64) {
	}, []*edb.Index{})

	Users = edb.AddTable(schema, "users", 1, func(row *m.User, ib *edb.IndexBuilder) {
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

	// Superadmins = edb.AddTable(schema, "superadmins", 1, func(row *m.Superadmin, ib *edb.IndexBuilder) {
	// 	ib.Add(SuperadminsByEmail, row.EmailNorm)
	// }, func(tx *edb.Tx, row *m.Superadmin, oldVer uint64) {
	// }, []*edb.Index{
	// 	SuperadminsByEmail,
	// })
	// SuperadminsByEmail = edb.AddIndex[string]("by_email")

	Waitlisters = edb.AddTable(schema, "waitlisters", 1, func(row *m.Waitlister, ib *edb.IndexBuilder) {
		ib.Add(WaitlistersByEmail, row.EmailNorm)
	}, func(tx *edb.Tx, row *m.Waitlister, oldVer uint64) {
	}, []*edb.Index{
		WaitlistersByEmail,
	})
	WaitlistersByEmail = edb.AddIndex[string]("by_email")

	Sessions = edb.AddTable(schema, "sessions", 1, func(row *m.Session, ib *edb.IndexBuilder) {
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
)
