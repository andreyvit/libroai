package mvp

import (
	"github.com/andreyvit/buddyd/mvp/mvpjobs"
	"github.com/andreyvit/buddyd/mvp/mvprpc"
)

type MethodImpl struct {
	*mvprpc.Method
	Call func(rc *RC, in any) (any, error)
}

type MethodRegistry struct {
	Methods map[string]*MethodImpl
	Jobs    map[*mvpjobs.Kind]*JobConfig
}

func (app *App) doCall(rc *RC, m *MethodImpl, in any) (any, error) {
	var out any
	callErr := app.InTx(rc, m.StoreAffinity, func() error {
		var err error
		out, err = m.Call(rc, in)
		return err
	})
	return out, callErr
}
