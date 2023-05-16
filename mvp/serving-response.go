package mvp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Redirect struct {
	Path   string
	Status int
	Values url.Values
}

type RawOutput struct {
	Data       []byte
	ContenType string
}

// DebugOutput can be returned by request handlers
type DebugOutput string

type ResponseHandled struct{}

func (app *App) writeResponse(rc *RC, output any, route *Route, w http.ResponseWriter, r *http.Request) error {
	for _, cookie := range rc.SetCookies {
		http.SetCookie(w, cookie)
	}
	switch output := output.(type) {
	case *ViewData:
		if output.View == "" {
			output.View = strings.ReplaceAll(route.routeName, ".", "-")
		}
		output.Route = route
		// output.SiteData = app.siteData
		output.App = app
		output.RC = rc
		b, err := app.Render(rc, output)
		if err != nil {
			return err
		}
		w.Write(b)
	case *Redirect:
		path := output.Path
		if len(output.Values) > 0 {
			path = path + "?" + output.Values.Encode()
		}
		if output.Status == 0 {
			output.Status = http.StatusSeeOther
		}
		http.Redirect(w, r, path, http.StatusSeeOther)
	case DebugOutput:
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(output))
	case ResponseHandled:
		break
	default:
		panic(fmt.Errorf("%s: invalid return value %T %v", route.desc, output, output))
	}
	return nil
}
