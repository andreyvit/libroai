package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
)

var schema = edb.NewSchema(edb.SchemaOpts{})

var (
	UserSignInAttempts = edb.AddTable[m.UserSignInAttempt](schema, "UserSignInAttempt", 1, func(row *m.UserSignInAttempt, ib *edb.IndexBuilder) {
	}, func(tx *edb.Tx, row *m.UserSignInAttempt, oldVer uint64) {
	}, []*edb.Index{})
)
