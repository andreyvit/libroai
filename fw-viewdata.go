package main

import (
	"fmt"
	"html/template"
)

type ViewData struct {
	View   string
	Title  string
	Layout string
	Data   any

	*SiteData
	Route *routeInfo
	App   *App

	// Content is only populated in layouts and contains the rendered content of the page
	Content template.HTML
	// Head is extra HEAD content
	Head template.HTML
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
	return &RenderData{
		Data:     value,
		Args:     m,
		ViewData: d.ViewData,
	}
}
