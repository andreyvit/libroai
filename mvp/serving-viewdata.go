package mvp

import (
	"fmt"
	"html/template"
	"strings"
)

type ViewData struct {
	View         string
	Title        string
	Layout       string
	Data         any
	SemanticPath string

	*SiteData
	Route *Route
	App   *App

	// Content is only populated in layouts and contains the rendered content of the page
	Content template.HTML
	// Head is extra HEAD content
	Head template.HTML
}

func (vd *ViewData) DefaultPathParams() map[string]string {
	defaults := make(map[string]string)
	return defaults
}

func (vd *ViewData) IsActive(path string) bool {
	return vd.SemanticPath == path || strings.HasPrefix(vd.SemanticPath, path+"/")
}

type SiteData struct {
	AppName string
}

type RenderData struct {
	Data any
	Args map[string]any
	*ViewData
}

func (d *RenderData) Bind(value any, args ...any) *RenderData {
	n := len(args)
	if n%2 != 0 {
		panic(fmt.Errorf("odd number of arguments %d: %v", n, args))
	}
	m := make(map[string]any, n/2)
	for i := 0; i < n; i += 2 {
		key, value := args[i], args[i+1]
		if keyStr, ok := key.(string); ok {
			m[keyStr] = value
		} else {
			panic(fmt.Errorf("argument %d must be a string, got %T: %v", i, key, key))
		}
	}
	if len(m) == 0 {
		m["__dummy"] = true
	}
	return &RenderData{
		Data:     value,
		Args:     m,
		ViewData: d.ViewData,
	}
}

func (d *RenderData) Value(name string) (any, bool) {
	v, found := d.Args[name]
	return v, found
}

func (d *RenderData) String(name string) (string, bool) {
	v, found := d.Value(name)
	if found {
		return stringifyArg(v), true
	}
	return "", false
}

func (d *RenderData) PopString(name string) (string, bool) {
	v, found := d.String(name)
	if found {
		delete(d.Args, name)
	}
	return v, found
}

func isPassThruArg(k string) bool {
	return passThruArgs[k] || strings.HasPrefix(k, "data-")
}

var passThruArgs = map[string]bool{
	"id":     true,
	"target": true,
	"rel":    true,
}

func stringifyArg(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}
