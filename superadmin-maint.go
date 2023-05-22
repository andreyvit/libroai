package main

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/andreyvit/buddyd/internal/httperrors"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/buddyd/mvp/forms"
)

type Procedure struct {
	Slug       string
	Title      string
	Form       *forms.Form
	RenderForm func() template.HTML
	Handler    func(rc *RC) error
}

func (app *App) Procedures() []*Procedure {
	return []*Procedure{
		app.importProcedure(),
	}
}

func (app *App) listSuperadminProcedures(rc *mvp.RC, in *struct{}) (*mvp.ViewData, error) {
	procs := app.Procedures()
	slugs := make(map[string]bool)
	for _, proc := range procs {
		if slugs[proc.Slug] {
			panic(fmt.Errorf("duplicate slug %s", proc.Slug))
		}
		slugs[proc.Slug] = true

		proc.Form.URL = app.URL("superadmin.maintenance.run", "procedure", proc.Slug)
		proc.Form.Group.Children = append(proc.Form.Group.Children, &forms.Wrapper{
			Template: "buttonbar-compact",
			Child: forms.Children{
				&forms.Button{
					TagOpts: forms.TagOpts{},
					Title:   proc.Title,
				},
			},
		})
		if proc.RenderForm == nil {
			proc := proc
			proc.RenderForm = func() template.HTML {
				return app.RenderForm(rc, proc.Form)
			}
		}
	}
	return &mvp.ViewData{
		View:         "superadmin/maintenance",
		Title:        "Maintenance Procedures",
		SemanticPath: "superadmin/maintenance",
		Data: struct {
			Procedures []*Procedure
		}{
			Procedures: procs,
		},
	}, nil
}

func (app *App) runSuperadminProcedure(rc *RC, in *struct {
	Slug string `form:"procedure,path" json:"-"`
}) (*mvp.ViewData, error) {
	var proc *Procedure
	for _, p := range app.Procedures() {
		if p.Slug == in.Slug {
			proc = p
			break
		}
	}
	if proc == nil {
		return nil, httperrors.Errorf(404, "", "Procedure not found")
	}

	var procErr error
	var logs strings.Builder
	if rc.HandleForm(proc.Form) {
		rc.LogTo(&logs)
		procErr = proc.Handler(rc)
		rc.LogTo(nil)
	} else {
		procErr = fmt.Errorf("validation error")
	}

	var message string
	if procErr == nil {
		message = "Succeeded."
	} else {
		message = fmt.Sprintf("Failed: %v", procErr)
	}

	return &mvp.ViewData{
		View:         "superadmin/maintenance-result",
		Title:        proc.Title,
		SemanticPath: "superadmin/maintenance",
		Data: struct {
			Procedure *Procedure
			OK        bool
			Message   string
			Logs      string
		}{
			Procedure: proc,
			OK:        procErr == nil,
			Message:   message,
			Logs:      logs.String(),
		},
	}, nil
}
