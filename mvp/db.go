package mvp

import (
	"log"
	"os"
	"path/filepath"

	"github.com/andreyvit/buddyd/mvp/flake"
	"github.com/andreyvit/edb"
)

func initAppDB(app *App, opt *AppOptions) {
	app.gen = flake.NewGen(0, 0)

	if app.Settings.DataDir == "" {
		app.Settings.DataDir = must(os.MkdirTemp("", "testdb*"))
	}
	ensure(os.MkdirAll(app.Settings.DataDir, 0755))
	app.db = must(edb.Open(filepath.Join(app.Settings.DataDir, "bolt.db"), app.Settings.Configuration.Schema, edb.Options{
		Logf:      log.Printf,
		Verbose:   false,
		IsTesting: false,
	}))
}

func closeAppDB(app *App) {
	app.db.Close()
}

func (app *App) NewID() flake.ID {
	return app.gen.New()
}

func (app *App) InTx(rc *RC, writable bool, f func() error) error {
	if rc.tx != nil {
		if writable && !rc.tx.IsWritable() {
			panic("cannot initiate a mutating tx from read-only one")
		}
		return f()
	} else {
		return app.db.Tx(writable, func(tx *edb.Tx) error {
			rc.tx = tx
			defer func() {
				rc.tx = nil
			}()
			return f()
		})
	}
}

func (app *App) Read(rc *RC, f func() error) error {
	return app.InTx(rc, false, f)
}
func (app *App) Write(rc *RC, f func() error) error {
	return app.InTx(rc, true, f)
}
func (app *App) MustRead(rc *RC, f func()) {
	ensure(app.InTx(rc, false, func() error {
		f()
		return nil
	}))
}
func (app *App) MustWrite(rc *RC, f func()) {
	ensure(app.InTx(rc, true, func() error {
		f()
		return nil
	}))
}
