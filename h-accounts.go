package main

import (
	"time"

	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/edb"
)

const (
	CodeSentNew      = 1
	CodeSentAgain    = 2
	CodeSentNotAgain = 3
)

func (app *App) showSignIn(rc *mvp.RC, in *struct {
	Email    string `json:"email"`
	CodeSent int    `json:"code_sent"`
	ErrorMsg string `json:"msg"`
}) (*mvp.ViewData, error) {
	return &mvp.ViewData{
		View:  "accounts/signin",
		Title: "Sign In",
		Data: struct {
			Email    string
			CodeSent int
			ErrorMsg string
		}{
			Email:    in.Email,
			CodeSent: in.CodeSent,
			ErrorMsg: "",
		},
	}, nil
}

func (app *App) handleSignIn(rc *mvp.RC, in *struct {
	IsSaving bool   `json:"-" form:",issave"`
	Email    string `json:"email"`
	Code     string `json:"code"`
}) (*mvp.ViewData, error) {
	if in.Email == "" {
		// message here
	}
	if in.Email != "" {
		a := edb.Get[m.UserSignInAttempt](rc, in.Email)
		if a == nil {
			a = &m.UserSignInAttempt{
				Email: in.Email,
			}
		}

		if in.Code != "" {

		}

		a.Code = mvp.RandomDigits(6)
		a.Time = rc.Now
		edb.Put(rc, a)

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
		return app.Redirect(0, "signin"), nil
	}

	return &mvp.ViewData{
		View:  "test",
		Title: "Test Page",
		Data: struct {
			Email    string
			CodeSent bool
			ErrorMsg string
		}{
			Email:    in.Email,
			CodeSent: in.CodeSent,
			ErrorMsg: "",
		},
	}, nil
}
