package mvp

import (
	"github.com/andreyvit/buddyd/mvp/mvpjobs"
	mvpm "github.com/andreyvit/buddyd/mvp/mvpmodel"
	"github.com/andreyvit/edb"
)

type Module struct {
	Name string

	SetupHooks  func(app *App)
	LoadSecrets func(*Settings, Secrets)

	DBSchema *edb.Schema
	Jobs     *mvpjobs.Schema
	Types    map[mvpm.Type][]string
}
