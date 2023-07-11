package main

import (
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/edb"
)

func loadRuntimeAccount(rc *RC, accountID m.AccountID) *m.RuntimeAccount {
	acc := edb.Get[m.Account](rc, accountID)
	if acc == nil {
		return nil
	}
	racc := &m.RuntimeAccount{
		Account:   acc,
		UsersByID: make(map[m.UserID]*m.User),
	}

	users := edb.All(edb.ExactIndexScan[m.User](rc, UsersByAccount, accountID))
	for _, u := range users {
		racc.UsersByID[u.ID] = u
	}

	return racc
}
