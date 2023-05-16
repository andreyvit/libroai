package main

import (
	"html/template"

	"github.com/andreyvit/buddyd/mvp"
)

func (app *App) registerViewHelpers(m template.FuncMap) {
	m["mood_text_class"] = func(mood mvp.Mood) string {
		switch mood {
		case mvp.MoodSuccess:
			return "text-green-700"
		case mvp.MoodFailure:
			return "text-red-600"
		case mvp.MoodSubtle:
			return "text-gray-500"
		default:
			return "text-gray-900"
		}
	}
}
