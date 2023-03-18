package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/andreyvit/buddyd/internal/httperrors"
	"github.com/uptrace/bunrouter"
)

type router struct {
	*App
	*bunrouter.Group
}

type routeInfo struct {
	FullName   string
	CallName   string
	FuncVal    reflect.Value
	InType     reflect.Type
	Idempotent bool
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
	rcPtrType reflect.Type = reflect.TypeOf((*RC)(nil))
	errorType reflect.Type = reflect.TypeOf((*error)(nil)).Elem()
)

func (router *router) Add(callName string, methodAndPath string, f any) {
	method, path, ok := strings.Cut(methodAndPath, " ")
	if !ok {
		panic(fmt.Errorf(`%s: %q is not a valid "METHOD path" string`, callName, methodAndPath))
	}
	mi, ok := validHTTPMethods[method]
	if !ok {
		panic(fmt.Errorf(`%s: %q is not a valid "METHOD path" string, method %q is invalid`, callName, methodAndPath, method))
	}

	fn := callName + " " + methodAndPath

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
		CallName:   callName,
		FuncVal:    fv,
		InType:     inTyp,
		Idempotent: mi.Idempotent,
	}

	app := router.App

	router.Group.Handle(method, path, func(w http.ResponseWriter, req bunrouter.Request) error {
		rc := app.NewHTTPRequestRC(req.Request)
		defer rc.Close()

		err := app.callRoute(route, rc, w, req)
		logRequest(rc, req.Request, err)
		if err != nil {
			http.Error(w, err.Error(), httperrors.HTTPCode(err))
		}
		return nil
	})
}
