package main

import (
	"fmt"
	"html/template"

	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	mvpm "github.com/andreyvit/buddyd/mvp/mvpmodel"
)

func (app *App) registerViewHelpers(funcs template.FuncMap) {
	funcs["can"] = func(data *mvp.RenderData, permStr string, objs ...any) bool {
		var obj mvpm.Object
		if len(objs) == 1 {
			obj = objs[0].(mvpm.Object)
		} else if len(objs) != 0 {
			panic(fmt.Errorf("invalid call of {{can}}"))
		}

		perm := must(m.ParsePermission(permStr))
		rc := fullRC.From(data.RC)
		return rc.Can(perm, obj)
	}
	funcs["mood_text_class"] = func(mood mvp.Mood) string {
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
