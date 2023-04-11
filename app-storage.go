package main

func (app *App) InTx(rc *RC, writable bool, f func() error) error {
	// if rc.Tx != nil {
	// 	if writable && !rc.Tx.IsWritable() {
	// 		panic("cannot initiate a mutating tx from read-only one")
	// 	}
	// 	return f()
	// } else {
	// 	return app.DB.Tx(writable, func(tx *db.Tx) error {
	// 		rc.Tx = tx
	// 		defer func() {
	// 			rc.Tx = nil
	// 		}()
	// 		return f()
	// 	})
	// }
	return f()
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
