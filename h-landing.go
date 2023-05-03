package main

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/andreyvit/buddyd/mvp"
)

func (app *App) showLandingHome(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	return &mvp.ViewData{
		View:  "home",
		Title: "LibroAI",
		Data:  struct{}{},
	}, nil
}

func (app *App) showWaitlist(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	return &mvp.ViewData{
		View:  "waitlist",
		Title: "You are on LibroAI waitlist",
		Data:  struct{}{},
	}, nil
}

func (app *App) handleLandingSignup(rc *mvp.RC, in *struct {
	Email       string `json:"email"`
	CompanyName string `json:"organization"`
	FullName    string `json:"name"`
}) (*mvp.Redirect, error) {
	app.SendEmail(rc, &mvp.Email{
		From:    "libroai@tarantsov.com",
		To:      "andrey@tarantsov.com",
		ReplyTo: in.Email,
		Subject: fmt.Sprintf("[LibroAI Signup] %s from %s", in.FullName, in.CompanyName),
		View:    "emails/signup",
		Data: map[string]any{
			"Email":       in.Email,
			"CompanyName": in.CompanyName,
			"FullName":    in.FullName,
			"Now":         time.Now().UTC().Format("2006-01-02 15:04"),
		},
		Category: "signupform",
	})

	return app.Redirect(0, "landing.waitlist"), nil
}

func runTemplate(code string, values map[string]any) string {
	t := template.New("")
	_, err := t.Parse(code)
	if err != nil {
		panic(fmt.Errorf("error parsing template: %v", err))
	}

	var buf strings.Builder
	err = t.Execute(&buf, values)
	if err != nil {
		panic(fmt.Errorf("error executing template: %v", err))
	}
	return buf.String()
}
