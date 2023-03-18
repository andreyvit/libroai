package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/uptrace/bunrouter"
)

//go:embed static
var embeddedStaticAssetsFS embed.FS

var staticAssetsFS fs.FS

func pickStaticFS(embedFS embed.FS, dir string) fs.FS {
	if serveAssetsFromDisk {
		_, file, _, _ := runtime.Caller(0)
		if file == "" {
			panic("missing source file path in binary")
		}
		return os.DirFS(filepath.Join(filepath.Dir(file), dir))
	} else {
		return must(fs.Sub(embedFS, dir))
	}
}

func initializeEmbeddedStatics() {
	staticAssetsFS = pickStaticFS(embeddedStaticAssetsFS, "static")
}

func setupStaticServer(bun *bunrouter.Router, urlPrefix string, f fs.FS) {
	h := http.FileServer(http.FS(f))
	h = http.StripPrefix(urlPrefix, h)

	bun.GET(urlPrefix+"/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
		markPrivateMutable(w)
		h.ServeHTTP(w, req.Request)
		return nil
	})
}
