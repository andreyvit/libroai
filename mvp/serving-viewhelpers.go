package mvp

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"

	"golang.org/x/exp/slices"
)

func (app *App) registerBuiltinViewHelpers(m template.FuncMap) {
	m["c_link"] = app.renderLink
	m["c_svg"] = app.renderSVG
	m["attr"] = Attr
	m["error"] = func(text string) template.HTML {
		panic(fmt.Errorf("%s", text))
	}
	m["eval"] = app.EvalTemplate
	m["iif"] = func(cond any, trueVal any, falseVal any) any {
		if IsFalsy(cond) {
			return falseVal
		} else {
			return trueVal
		}
	}
	m["url_for"] = func(d *RenderData, name string, extras ...any) template.URL {
		defaults := d.DefaultPathParams()
		if len(defaults) > 0 {
			newExtras := make([]any, 0, len(extras)+1)
			newExtras = append(newExtras, defaults)
			newExtras = append(newExtras, extras...)
			extras = newExtras
		}
		return template.URL(d.App.URL(name, extras...))
	}
	m["repeat"] = func(n int) []int {
		r := make([]int, 0, n)
		for i := 0; i < n; i++ {
			r = append(r, i)
		}
		return r
	}
	m["pick"] = func(i int, values ...any) any {
		return values[i%len(values)]
	}
	m["picks"] = func(s string, values ...any) any {
		for i := 0; i < len(values)-1; i += 2 {
			if i+1 >= len(values) {
				// "else" clause
				return values[i]
			}
			key := fmt.Sprint(values[i])
			if key == s {
				return values[i+1]
			}
		}
		return nil
	}
}

func (app *App) renderLink(data *RenderData) template.HTML {
	classes := make([]string, 0, 8)

	href, _ := data.PopString("href")
	routeName, _ := data.PopString("route")
	bodyAttr, _ := data.PopString("body")
	classAttr, _ := data.PopString("class")
	activeClassAttr, _ := data.PopString("active-class")
	if activeClassAttr == "" {
		activeClassAttr = "active"
	}
	sempathAttr, _ := data.PopString("sempath")
	iconAttr, _ := data.PopString("icon")

	var isActive, looksActive bool
	if href != "" {
		if sempathAttr != "" {
			looksActive = data.IsActive(sempathAttr)
			isActive = looksActive
		}
	} else if routeName != "" {
		route := app.routesByName[routeName]
		if route == nil {
			panic(fmt.Errorf("unknown route %s", routeName))
		}

		params := data.DefaultPathParams()
		for _, k := range route.pathParams {
			v, found := data.PopString(k)
			if found {
				params[k] = v
			} else if _, found = params[k]; !found {
				panic(fmt.Errorf("route %s requires path param %s", routeName, k))
			}
		}

		href = app.URL(routeName, params)
		isActive = (data.Route.routeName == routeName)
		looksActive = isActive
		if sempathAttr != "" {
			looksActive = data.IsActive(sempathAttr)
		}
	}

	var iconStr template.HTML
	if iconAttr != "" {
		iconStr = app.renderSVG(data.Bind(nil, "class", "icon", "src", iconAttr))
		classes = append(classes, "with-icon")
	}

	classes = append(classes, strings.Fields(classAttr)...)

	if looksActive && activeClassAttr != "" {
		for _, c := range strings.Fields(activeClassAttr) {
			if s, ok := strings.CutPrefix(c, "remove:"); ok {
				if i := slices.Index(classes, s); i >= 0 {
					classes = slices.Delete(classes, i, i+1)
				}
			} else {
				if i := slices.Index(classes, s); i < 0 {
					classes = append(classes, c)
				}
			}
		}
	}

	var extraArgs strings.Builder
	for k, v := range data.Args {
		if isPassThruArg(k) {
			extraArgs.WriteString(string(Attr(k, v)))
		} else {
			panic(fmt.Errorf("<c-link>: invalid param %s", k))
		}
	}
	if len(classes) > 0 {
		extraArgs.WriteString(string(Attr("class", strings.Join(classes, " "))))
	}

	if isActive || href == "" {
		return template.HTML(fmt.Sprintf(`<div%s>%s%s</div>`, extraArgs.String(), iconStr, template.HTMLEscapeString(bodyAttr)))
	} else {
		return template.HTML(fmt.Sprintf(`<a href="%s"%s>%s%s</a>`, href, extraArgs.String(), iconStr, template.HTMLEscapeString(bodyAttr)))
	}
}

func (app *App) renderSVG(data *RenderData) template.HTML {
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

	raw, err := fs.ReadFile(data.App.StaticFS(), src)
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
	return template.HTMLAttr(" " + name + "=\"" + template.HTMLEscapeString(fmt.Sprint(value)) + "\"")
}
