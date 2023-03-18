package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/uptrace/bunrouter"
)

type Redirect struct {
	Path   string
	Values url.Values
}

type RawOutput struct {
	Data       []byte
	ContenType string
}

// DebugOutput can be returned by request handlers
type DebugOutput string

func (app *App) returnResponse(rc *RC, output any, route *routeInfo, w http.ResponseWriter, req bunrouter.Request) error {
	switch output := output.(type) {
	case *ViewData:
		if output.View == "" {
			output.View = strings.ReplaceAll(route.CallName, ".", "_")
		}
		output.SiteData = app.siteData
		output.CallName = route.CallName
		output.FullCallName = route.FullName
		b, err := app.render(rc, output)
		if err != nil {
			return err
		}
		w.Write(b)
	case Redirect:
		path := output.Path
		if len(output.Values) > 0 {
			path = path + "?" + output.Values.Encode()
		}
		http.Redirect(w, req.Request, path, http.StatusSeeOther)
	case DebugOutput:
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(output))
	default:
		panic(fmt.Errorf("%s: invalid return value %T %v", route.FullName, output, output))
	}
	return nil
}
