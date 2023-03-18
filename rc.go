package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

// RC stands for Request Context.
type RC struct {
	Ctx context.Context
	// Auth      bm.Auth
	RequestID string
	Start     time.Time // ACTUAL time of request start
	Now       time.Time // wall clock time of request start
	logf      func(format string, args ...any)
}

func (app *App) NewHTTPRequestRC(r *http.Request) *RC {
	return app.NewRC(r.Context(), r.Header.Get("X-Request-ID"))
}

func (app *App) NewRC(ctx context.Context, requestID string) *RC {
	if requestID == "" {
		requestID = randomHex(32)
	}
	return &RC{
		Ctx:       ctx,
		RequestID: requestID,
		Start:     time.Now(),
		Now:       time.Now(),
		logf:      app.Logf,
	}
}

func (rc *RC) Context() context.Context {
	return rc.Ctx
}

func (rc *RC) Logf(format string, args ...any) {
	rc.logf(format, args...)
}

func (rc *RC) AppendLogPrefix(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "[%s] ", rc.RequestID)
}
func (rc *RC) AppendLogSuffix(buf *bytes.Buffer) {
}

func (rc *RC) Close() {
}
