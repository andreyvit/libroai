package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"strings"

	"github.com/andreyvit/buddyd/internal/flogger"
	"github.com/andreyvit/minicomponents"
)

//go:embed views/*.html
var embeddedViewsFS embed.FS

type templKind int

const (
	partialTempl = templKind(iota)
	componentTempl
	layoutTempl
	pageTempl
)

func (app *App) render(lc flogger.Context, data *ViewData) ([]byte, error) {
	if data.Layout == "" {
		data.Layout = "default"
	}

	t := app.templates
	if settings.ServeAssetsFromDisk {
		flogger.Log(lc, "reloading templates from disk")
		var err error
		t, err = app.loadTemplates()
		if err != nil {
			panic(fmt.Errorf("reloading templates: %v", err))
		}
	}

	rdata := &RenderData{Data: data.Data, ViewData: data}

	var buf strings.Builder
	err := t.ExecuteTemplate(&buf, data.View, rdata)
	if err != nil {
		return nil, err
	}
	data.Content = template.HTML(buf.String())

	if data.Layout == "none" {
		return []byte(data.Content), nil
	}

	var buf2 bytes.Buffer
	err = t.ExecuteTemplate(&buf2, "layout-"+data.Layout, rdata)
	if err != nil {
		return nil, err
	}
	return buf2.Bytes(), nil
}

type templDef struct {
	name string
	path string
	code string
	kind templKind
	tmpl *template.Template
}

func (app *App) loadTemplates() (*template.Template, error) {
	const templateSuffix = ".html"
	viewsFS := pickEmbeddedFS(embeddedViewsFS, "views")

	funcs := make(template.FuncMap, 100)
	app.registerViewHelpers(funcs)

	root := template.New("")
	root.Funcs(funcs)

	var templs []*templDef

	err := fs.WalkDir(viewsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if !strings.HasSuffix(path, templateSuffix) {
			return nil
		}
		name := strings.TrimSuffix(d.Name(), templateSuffix)
		code := string(must(fs.ReadFile(viewsFS, path)))

		var kind templKind
		if strings.HasPrefix(name, "c-") {
			kind = componentTempl
		} else if strings.HasPrefix(name, "layout-") {
			kind = layoutTempl
		} else if strings.Contains(name, "__") {
			kind = partialTempl
		} else {
			kind = pageTempl
		}

		templs = append(templs, &templDef{
			name: name,
			path: path,
			code: code,
			kind: kind,
			tmpl: root.New(name),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	comps := make(map[string]*minicomponents.ComponentDef)
	for _, tmpl := range templs {
		if tmpl.kind == componentTempl {
			comps[tmpl.name] = &minicomponents.ComponentDef{
				RenderMethod: minicomponents.RenderMethodTemplate,
			}
		}
	}
	for k := range funcs {
		if strings.HasPrefix(k, "c_") {
			name := strings.ReplaceAll(k, "_", "-")
			comps[name] = &minicomponents.ComponentDef{
				RenderMethod: minicomponents.RenderMethodFunc,
				FuncName:     k,
			}
		}
	}

	for _, tmpl := range templs {
		code := tmpl.code
		code, _ = minicomponents.Rewrite(code, tmpl.name, comps)

		if tmpl.kind == componentTempl {
			code = "{{with .Args}}" + code + "{{end}}"
		} else if tmpl.kind == pageTempl || tmpl.kind == partialTempl {
			code = "{{with .Data}}" + code + "{{end}}"
		}
		_, err = tmpl.tmpl.Parse(code)
		if err != nil {
			return nil, fmt.Errorf("error parsing %v: %w", tmpl.path, err)
		}
	}

	return root, nil
}
