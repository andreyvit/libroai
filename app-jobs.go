package main

import (
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/mvpjobs"
)

var (
	jobProduceAnswer = jobSchema.Define("ProduceAnswer", nil, mvpjobs.Repeatable, mvpjobs.Ephemeral)
)

func (app *App) registerJobs(b mvp.JobRegistry) {
}
