package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/andreyvit/buddyd/internal/httperrors"
	"github.com/andreyvit/httpform"
	"github.com/uptrace/bunrouter"
)

type urlOption int

const Absolute = urlOption(1)

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
		case urlOption:
			if extra == Absolute {
				abs = true
			}
		default:
			panic(fmt.Errorf("route %s: unsupported extra %T %v", name, extra, extra))
		}
	}

	path := route.Path
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

func (app *App) setupHandler() http.Handler {
	var web http.Handler
	{
		bun := bunrouter.New()
		g := routeBuilder{app: app, bg: &bun.Group}
		app.registerRoutes(&g)
		web = bun
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch trimPort(r.Host) {
		case app.BaseURL.Host:
			web.ServeHTTP(w, r)
		default:
			http.Error(w, fmt.Sprintf("Invalid domain %q (wanted %s)", trimPort(r.Host), app.BaseURL.Host), http.StatusMisdirectedRequest)
		}
	})
}

// callRoute handles an HTTP request using the middleware, handler and parameters of the given route.
//
// Note: we might be doing too much here, some logic should probably be moved into middleware.
func (app *App) callRoute(route *routeInfo, rc *RC, w http.ResponseWriter, req bunrouter.Request) error {
	rc.Route = route

	if !route.Idempotent {
		// TODO: check CSRF
	}

	inVal := reflect.New(route.InType)
	err := httpform.Default.DecodeVal(req.Request, req.Params(), inVal)
	if err != nil {
		return err
	}

	var output any

	err = app.InTx(rc, !route.Idempotent, func() error {
		for _, mw := range route.Middleware {
			if mw.f == nil {
				continue
			}
			// if mw.name != "" {
			// 	flogger.Log(rc, "middleware %s", mw.name)
			// }
			var err error
			output, err = mw.f(rc)
			if err != nil {
				return err
			}
			if output != nil {
				return nil
			}
		}

		inputs := make([]reflect.Value, route.FuncVal.Type().NumIn())
		inputs[0] = reflect.ValueOf(rc)
		inputs[1] = inVal

		results := route.FuncVal.Call(inputs)
		output = results[0].Interface()
		if errVal := results[1].Interface(); errVal != nil {
			return errVal.(error)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return app.writeResponse(rc, output, route, w, req.Request)
}

type routeInfo struct {
	FullName   string
	RouteName  string
	Method     string
	Path       string
	FuncVal    reflect.Value
	InType     reflect.Type
	Idempotent bool
	Middleware middlewareSlotList
	PathParams []string
}

// routeBuilder helps to define named routes, and holds the path and middleware.
type routeBuilder struct {
	app    *App
	bg     *bunrouter.Group
	mwlist middlewareSlotList
	path   string
}

// group defines a group of routes with a common prefix and/or middleware.
func (g *routeBuilder) group(path string, f func(g *routeBuilder)) {
	sg := routeBuilder{app: g.app, bg: g.bg.NewGroup(path), mwlist: g.mwlist.Clone(), path: g.path + path}
	f(&sg)
}

// use adds the given middleware. If the slot is already occupied, overrides the middleware
// in the given slot (so that you can define a default authentication or CORS middleware,
// but override it for certain groups or routes).
func (g *routeBuilder) use(slot string, middleware any) {
	g.mwlist.Add(slot, adaptMiddleware(middleware))
}

func (g *routeBuilder) static(path string, fs fs.FS) {
	setupStaticServer(g.bg, path, fs)
}

// route defines a named route. methodAndPath are space-separated.
//
// The handler function must have func(rc *RC, in *SomeStruct) (output, error) signature,
// where output can be anything, but must be something that app.writeResponse accepts.
func (g *routeBuilder) route(routeName string, methodAndPath string, f any, options ...func(route *routeInfo)) *routeInfo {
	method, path, ok := strings.Cut(methodAndPath, " ")
	if !ok {
		panic(fmt.Errorf(`%s: %q is not a valid "METHOD path" string`, routeName, methodAndPath))
	}
	mi, ok := validHTTPMethods[method]
	if !ok {
		panic(fmt.Errorf(`%s: %q is not a valid "METHOD path" string, method %q is invalid`, routeName, methodAndPath, method))
	}

	fn := routeName + " " + methodAndPath

	fv := reflect.ValueOf(f)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(fmt.Errorf(`%s: function is invalid, got %v, wanted a function`, fn, ft))
	}
	if ft.NumOut() != 2 || ft.Out(1) != errorType {
		panic(fmt.Errorf(`%s: got %v, wanted a function returning (something, error)`, fn, ft))
	}
	if ft.NumIn() != 2 || ft.In(0) != rcPtrType || ft.In(1).Kind() != reflect.Ptr || ft.In(1).Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf(`%s: got %v, wanted a function accepting (*RC, *SomeStruct)`, fn, ft))
	}
	// inTypPtr := ft.In(1)
	inTyp := ft.In(1).Elem()

	route := &routeInfo{
		FullName:   fn,
		RouteName:  routeName,
		Method:     method,
		Path:       g.path + path,
		FuncVal:    fv,
		InType:     inTyp,
		Idempotent: mi.Idempotent,
		Middleware: g.mwlist.Clone(),
	}

	for _, param := range pathParamsRe.FindAllString(route.Path, -1) {
		route.PathParams = append(route.PathParams, param[1:])
	}

	for _, f := range options {
		f(route)
	}

	if prev := g.app.routesByName[route.RouteName]; prev != nil {
		panic(fmt.Errorf("route %s: duplicate path %s, previous was %s", route.RouteName, methodAndPath, prev.Method+" "+prev.Path))
	}
	g.app.routesByName[route.RouteName] = route

	g.bg.Handle(method, path, func(w http.ResponseWriter, req bunrouter.Request) error {
		rc := g.app.NewHTTPRequestRC(w, req)
		defer rc.Close()

		err := g.app.callRoute(route, rc, w, req)
		logRequest(rc, req.Request, err)
		if err != nil {
			http.Error(w, err.Error(), httperrors.HTTPCode(err))
		}
		return nil
	})

	return route
}

// use specifies per-route middleware; see routeBuilder.use for more info.
func (route *routeInfo) use(slot string, f middlewareFunc) {
	route.Middleware.Add(slot, f)
}

// mutator is an option you can pass into routeBuilder.
func mutator(route *routeInfo) {
	route.Idempotent = false
}

var validHTTPMethods = map[string]struct {
	Idempotent bool
}{
	"GET":    {Idempotent: true},
	"POST":   {},
	"PUT":    {},
	"DELETE": {},
	"OPTION": {Idempotent: true},
}

var (
	rcPtrType    reflect.Type = reflect.TypeOf((*RC)(nil))
	errorType    reflect.Type = reflect.TypeOf((*error)(nil)).Elem()
	pathParamsRe              = regexp.MustCompile(`:(\w+)`)
)
