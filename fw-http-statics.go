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

func pickEmbeddedFS(embedFS embed.FS, dir string) fs.FS {
	if settings.ServeAssetsFromDisk {
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
	staticAssetsFS = pickEmbeddedFS(embeddedStaticAssetsFS, "static")
}

func setupStaticServer(g *bunrouter.Group, urlPrefix string, f fs.FS) {
	h := http.FileServer(http.FS(f))
	h = http.StripPrefix(urlPrefix, h)

	g.GET(urlPrefix+"/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
		markPrivateMutable(w)
		h.ServeHTTP(w, req.Request)
		return nil
	})
}
