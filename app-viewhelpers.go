package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"
)

func (app *App) registerViewHelpers(m template.FuncMap) {
	m["c_svg"] = renderSVG
	m["attr"] = Attr
	m["prefix_classes"] = PrefixClasses
}

func IsFalsy(value any) bool {
	return value == nil || value == "" || value == false
}

func Attr(name string, value any) template.HTMLAttr {
	if IsFalsy(value) {
		return ""
	}
	if value == true {
		return template.HTMLAttr(" " + name)
	}
	return template.HTMLAttr(" " + name + "=" + template.HTMLEscapeString(fmt.Sprint(value)))
}

func PrefixClasses(prefix string, classes any) string {
	if IsFalsy(classes) {
		return ""
	}
	items := strings.Fields(fmt.Sprint(classes))
	for i, item := range items {
		items[i] = prefix + item
	}
	return strings.Join(items, " ")
}

func renderSVG(data *RenderData) template.HTML {
	var src string
	var extraArgs strings.Builder
	for k, v := range data.Args {
		if k == "src" {
			src = fmt.Sprint(v)
		} else {
			extraArgs.WriteString(string(Attr(k, v)))
		}
	}
	if src == "" {
		panic("<c-svg>: missing src attribute")
	}

	raw, err := fs.ReadFile(staticAssetsFS, src)
	if err != nil {
		// var got []string
		// ensure(fs.WalkDir(app.staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		// 	got = append(got, path)
		// 	return nil
		// }))
		// panic(fmt.Errorf("<c-svg>: cannot find static/%s, have: %s", src, strings.Join(got, ", ")))
		panic(fmt.Errorf("<c-svg>: cannot find %s under static/", src))
	}
	body := string(raw)

	if extraArgs.Len() > 0 {
		i := strings.Index(body, "<svg ")
		if i < 0 {
			panic(fmt.Errorf("<c-svg>: static/%s: cannot find <svg> opening tag", src))
		}

		body = body[:i+4] + extraArgs.String() + body[i+4:]
	}
	return template.HTML(body)
}
