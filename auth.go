package main

import (
	"fmt"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/flogger"
	mvpm "github.com/andreyvit/mvp/mvpmodel"

	m "github.com/andreyvit/buddyd/model"
)

func resetAuth(app *App, rc *RC) {
	rc.Session = nil
	rc.OriginalUser = nil
	rc.User = nil
	rc.Account = nil
}

func loadSessionAndUser(app *App, rc *RC) error {
	auth := rc.Auth()
	if auth.SessionID != 0 {
		rc.Session = edb.Get[m.Session](rc, auth.SessionID)
		if rc.Session == nil {
			return fmt.Errorf("session no longer exists")
		}
	}
	if auth.ActorRef.Type == mvpm.TypeUser {
		rc.OriginalUser = edb.Get[m.User](rc, auth.ActorRef.ID)
		if rc.OriginalUser == nil {
			return fmt.Errorf("user no longer exists")
		}
	}
	if rc.Session != nil && rc.Session.ImpersonatedUserID != 0 {
		rc.User = edb.Get[m.User](rc, rc.Session.ImpersonatedUserID)
		if rc.User == nil {
			rc.User = rc.OriginalUser
		}
	} else {
		rc.User = rc.OriginalUser
	}
	if rc.Session != nil && rc.OriginalUser != nil && rc.Session.AccountID != 0 {
		rc.Account = loadRuntimeAccount(rc, rc.Session.AccountID)
		if rc.Account == nil {
			rc.Session.AccountID = 0
			flogger.Log(rc, "WARNING: session refers to non-existent account %v", rc.Session.AccountID)
		}
	}
	return nil
}

func requireSuperadmin(rc *RC) (any, error) {
	if !rc.IsLoggedIn() {
		return rc.App().Redirect("signin"), nil
	}
	if err := rc.Check(m.PermissionAccessSuperadminArea, nil); err != nil {
		return nil, mvp.ErrForbidden.Wrap(err)
	}
	return nil, nil
}

func requireAdmin(rc *RC) (any, error) {
	if !rc.IsLoggedIn() {
		return rc.App().Redirect("signin"), nil
	}
	if err := rc.Check(m.PermissionAccessAdminArea, nil); err != nil {
		return nil, mvp.ErrForbidden.Wrap(err)
	}
	return nil, nil
}

func requireLoggedIn(rc *RC) (any, error) {
	if !rc.IsLoggedIn() {
		return redirectToLogIn(rc), nil
	}
	return nil, nil
}

func redirectToLogIn(rc *RC) *mvp.Redirect {
	return rc.App().Redirect("signin")
}

// func (app *App) makeWebToken(sess *bm.Session, now time.Time) string {
// 	account := fmt.Sprintf("%s-%s-%s", sess.ActorType.KeyString(), sess.ActorID.String(), sess.ID.String())
// 	return app.webAdminTokens.SignAt(now, account)
// }

// func (app *App) Logout(rc *RC, path string) (*Redirect, error) {
// 	app.DeleteAuthCookie(rc)
// 	if path == "" {
// 		path = app.URL("home")
// 	}
// 	return &Redirect{
// 		Path: path,
// 	}, nil
// }

// func (app *App) startSession(rc *RC, actor bm.ActorRef) {
// 	sess := &bm.Session{
// 		ID:           app.NewID(),
// 		ActorType:    actor.ActorType,
// 		ActorID:      actor.ActorID,
// 		CreationTime: rc.Now,
// 		RefreshTime:  rc.Now,
// 		ActivityTime: rc.Now,
// 	}
// 	edb.Put(rc, sess)
// 	token := app.makeWebToken(sess, rc.Now)
// 	rc.SetCookie(makeWebTokenCookie(token, shopifyLoginValidity))
// }
