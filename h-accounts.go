package main

import (
	"crypto/subtle"
	"net/url"
	"strconv"
	"time"

	"github.com/andreyvit/buddyd/internal/flogger"
	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	mvpm "github.com/andreyvit/buddyd/mvp/mvpmodel"
	"github.com/andreyvit/edb"
)

const (
	CodeNone        = 0
	CodeSentOld     = 1
	CodeSentInitial = 2
	CodeResentNew   = 3
	CodeResentAgain = 4
)

func (app *App) showSignIn(rc *mvp.RC, in *struct {
	Email    string `json:"email"`
	CodeSent int    `json:"code_sent"`
	EmailErr string `json:"email_err"`
	CodeErr  string `json:"code_err"`
}) (*mvp.ViewData, error) {
	var emailMsg, codeMsg *mvp.Msg
	if in.EmailErr != "" {
		emailMsg = mvp.FailureMsg(in.EmailErr)
	}
	switch in.CodeSent {
	case CodeResentNew:
		if in.CodeErr != "" {
			codeMsg = mvp.FailureMsg(in.CodeErr + " " + "A new code has been sent to your email.")
		} else {
			codeMsg = mvp.SuccessMsg("A new code has been sent to your email.")
		}
	case CodeResentAgain:
		if in.CodeErr != "" {
			codeMsg = mvp.FailureMsg(in.CodeErr + " " + "The code has been sent to your email again.")
		} else {
			codeMsg = mvp.SuccessMsg("The code has been sent to your email again.")
		}
	case CodeNone:
		if in.CodeErr != "" {
			codeMsg = mvp.FailureMsg(in.CodeErr)
		}
	default:
		if in.CodeErr != "" {
			codeMsg = mvp.FailureMsg(in.CodeErr)
		} else {
			codeMsg = mvp.SubtleMsg("Enter the code you have received via email.")
		}
	}

	return &mvp.ViewData{
		View:   "accounts/signin",
		Title:  "Sign In",
		Layout: "bare",
		Data: struct {
			Email    string
			CodeSent int
			ErrorMsg string
			EmailMsg *mvp.Msg
			CodeMsg  *mvp.Msg
		}{
			Email:    in.Email,
			CodeSent: in.CodeSent,
			ErrorMsg: "",
			EmailMsg: emailMsg,
			CodeMsg:  codeMsg,
		},
	}, nil
}

func (app *App) handleSignIn(rc *mvp.RC, in *struct {
	IsSaving bool   `json:"-" form:",issave"`
	Email    string `json:"email"`
	Code     string `json:"code"`
	Resend   bool   `json:"resend"`
}) (any, error) {
	if in.Email == "" {
		return app.Redirect(0, "signin", url.Values{
			"email":     {in.Email},
			"email_err": {"Email is required."},
		}), nil
	}

	a := edb.Get[m.UserSignInAttempt](rc, in.Email)
	if a == nil {
		a = &m.UserSignInAttempt{
			Email: in.Email,
		}
	}

	if a.Code != "" {
		exp := a.Time.Add(app.Settings().SignInCodeExpiration.Value())
		if rc.Now.After(exp) {
			a.Code = ""
		}
	}

	var codeErr string
	if in.Code != "" && a.Code != "" {
		if 1 == subtle.ConstantTimeCompare([]byte(in.Code), []byte(a.Code)) {
			return app.finishSignIn(rc, a.Email)
		} else {
			codeErr = "Code is incorrect."
		}
	}

	var sent int
	if a.Code == "" {
		a.Code = mvp.RandomDigits(6)
		a.Time = rc.Now
		edb.Put(rc, a)
		if in.Resend {
			sent = CodeResentNew
		} else {
			sent = CodeSentInitial
		}
	} else {
		if rc.Now.Sub(a.Time) >= time.Duration(app.Settings().SignInCodeResendInterval) {
			sent = CodeResentAgain
		} else {
			sent = CodeSentOld
		}
	}

	if sent >= CodeSentInitial {
		app.SendEmail(rc, &mvp.Email{
			To:      in.Email,
			Subject: "LibroAI Sign In",
			View:    "emails/signin",
			Data: map[string]any{
				"Email": a.Email,
				"Code":  a.Code,
				"Now":   time.Now().UTC().Format("2006-01-02 15:04"),
			},
			Category: "signin",
		})
	}
	return app.Redirect(0, "signin", url.Values{
		"email":     {in.Email},
		"code_sent": {strconv.Itoa(sent)},
		"code_err":  {codeErr},
	}), nil
}

func (app *App) finishSignIn(rc *mvp.RC, email string) (any, error) {
	flogger.Log(rc, "Signed in as %s", email)
	emailNorm := mvp.CanonicalEmail(email)

	if u := edb.Lookup[m.User](rc, UsersByEmail, emailNorm); u != nil {
		app.startSession(rc, u)
		return app.openApp(fullRC.From(rc))
	}

	wl := edb.Lookup[m.Waitlister](rc, WaitlistersByEmail, emailNorm)
	if wl == nil {
		wl = &m.Waitlister{
			ID:        app.NewID(),
			Email:     email,
			EmailNorm: emailNorm,
			LastLogin: rc.Now,
		}
		edb.Put(rc, wl)
	}
	app.startSession(rc, wl)
	return app.Redirect(0, "landing.waitlist"), nil
}

func (app *App) openApp(rc *RC) (*mvp.Redirect, error) {
	if rc.Can(m.PermissionManageAccount, nil) {
		return rc.Redirect(0, "admin.home"), nil
	} else {
		return rc.Redirect(0, "chat.home"), nil
	}
}

func (app *App) startSession(rc *mvp.RC, actor m.Actor) {
	sess := &m.Session{
		ID:           app.NewID(),
		Actor:        mvpm.RefTo(actor),
		LastActivity: rc.Now,
	}
	if user, ok := actor.(*m.User); ok {
		if len(user.Memberships) > 0 {
			sess.AccountID = user.Memberships[0].AccountID
		}
	}
	edb.Put(rc, sess)
	rc.SetAuthUsingCookie(mvp.Auth{
		SessionID: sess.ID,
		ActorRef:  sess.Actor,
	})
}

func (app *App) showPickAccountForm(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	wls := edb.All(edb.TableScan[m.Waitlister](rc, edb.FullScan()))
	users := edb.All(edb.TableScan[m.User](rc, edb.FullScan()))

	return &mvp.ViewData{
		View:  "superadmin/home",
		Title: "Superadmin",
		Data: struct {
			Waitlisters []*m.Waitlister
			Users       []*m.User
		}{
			Waitlisters: wls,
			Users:       users,
		},
	}, nil
}

func (app *App) handleSignOut(rc *mvp.RC, in *struct{}) (any, error) {
	rc.DeleteAuthCookie()
	return &mvp.Redirect{
		Path: app.URL("signin"),
	}, nil
}
