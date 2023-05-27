package main

import (
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/buddyd/mvp/mvpjobs"
)

var (
	jobProduceAnswer = jobSchema.Define("ProduceAnswer", nil, mvpjobs.Repeatable, mvpjobs.Ephemeral)
)

func (app *App) registerJobs(b mvp.JobRegistry) {
}
