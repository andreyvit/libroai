package mvp

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

type URLGenOption int

const Absolute = URLGenOption(1)

// URL generates a relative (default) or absolute URL based on a named route.
// Supply map[string]string or pairs of (key string, value string) arguments
// to supply path parameters for the route.
// Supply url.Values to add query parameters.
// Supply Absolute flag to return an absolute URL.
func (app *App) URL(name string, extras ...any) string {
	route := app.routesByName[name]
	if route == nil {
		panic(fmt.Errorf("route %s does not exist", name))
	}

	var keys map[string]string
	var params url.Values
	var abs bool
	for i := 0; i < len(extras); i++ {
		switch extra := extras[i].(type) {
		case map[string]string:
			if keys == nil {
				keys = extra
			} else {
				for k, v := range extra {
					keys[k] = v
				}
			}
		case url.Values:
			params = extra
		case string:
			if i+1 >= len(extras) {
				panic(fmt.Errorf("route %s: no value following extra key %q", name, extra))
			}
			if keys == nil {
				keys = make(map[string]string)
			}
			keys[extra] = fmt.Sprint(extras[i+1])
			i++
		case URLGenOption:
			if extra == Absolute {
				abs = true
			}
		default:
			panic(fmt.Errorf("route %s: unsupported extra %T %v", name, extra, extra))
		}
	}

	path := route.path
	if keys != nil {
		for k, v := range keys {
			path = strings.ReplaceAll(path, ":"+k, v)
		}
	}
	if strings.Contains(path, ":") {
		panic(fmt.Errorf("route %s: not all path params specified in %q", name, path))
	}

	var p url.URL
	if abs {
		p = *app.BaseURL
	}
	p.Path = path
	if params != nil {
		p.RawQuery = params.Encode()
	}

	log.Printf("URL(%s, %v) = %s", name, extras, p.String())
	return p.String()
}

// Redirect returns a response object that redirects to a given route.
// Status code defaults to 303 See Other.
// Supply map[string]string or pairs of (key string, value string) arguments
// to supply path parameters for the route.
// Supply url.Values to add query parameters.
// Supply Absolute flag to use an absolute URL.
func (app *App) Redirect(statusCode int, name string, extras ...any) *Redirect {
	path := app.URL(name, extras...)
	return &Redirect{
		Path:   path,
		Status: statusCode,
	}
}
