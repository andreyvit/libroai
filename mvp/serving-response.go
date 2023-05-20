package mvp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Redirect struct {
	Path       string
	StatusCode int
	Values     url.Values
}

// SameMethod uses 307 Temporary Redirect for this redirect.
// Same HTTP method will be used for the redirected request,
// unlike the default 303 See Other response which always redirects with a GET.
func (redir *Redirect) SameMethod() *Redirect {
	redir.StatusCode = http.StatusTemporaryRedirect
	return redir
}

// Permanent uses 308 Permanent Redirect for this redirect.
// Same HTTP method will be used for the redirected request.
//
// Note that there is no permanent redirection code that is guaranteed to
// always use GET. 301 Moved Permanently may or may not do that and is not
// recommended.
func (redir *Redirect) Permanent() *Redirect {
	redir.StatusCode = http.StatusPermanentRedirect
	return redir
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
		app.fillViewData(output, rc)
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
		code := output.StatusCode
		if code == 0 {
			code = http.StatusSeeOther
		}
		http.Redirect(w, r, path, code)
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

func (app *App) fillViewData(output *ViewData, rc *RC) {
	output.RC = BaseRC.AnyFull(rc)
	output.baseRC = rc
	output.App = app
}
