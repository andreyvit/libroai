package main

import (
	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"

	m "github.com/andreyvit/buddyd/model"
)

type AccountVM struct {
	*m.Account
	IsCurrent bool
}

func (app *App) listSuperadminAccounts(rc *RC, in *struct{}) (*mvp.ViewData, error) {
	wls := edb.All(edb.TableScan[m.Waitlister](rc, edb.FullScan()))
	rawAccounts := edb.All(edb.TableScan[m.Account](rc, edb.FullScan()))
	curAccountID := rc.AccountID()

	accounts := make([]*AccountVM, 0, len(rawAccounts))
	for _, acc := range rawAccounts {
		if acc != nil {
			accounts = append(accounts, &AccountVM{
				Account:   acc,
				IsCurrent: acc.ID == curAccountID,
			})
		}
	}

	return &mvp.ViewData{
		View:         "superadmin/accounts",
		Title:        "Accounts",
		SemanticPath: "superadmin/accounts",
		Data: struct {
			Waitlisters []*m.Waitlister
			Accounts    []*AccountVM
		}{
			Waitlisters: wls,
			Accounts:    accounts,
		},
	}, nil
}
