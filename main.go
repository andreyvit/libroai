package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/andreyvit/envloader"
	"github.com/andreyvit/httpserver"
	"github.com/andreyvit/openai"

	"github.com/andreyvit/buddyd/internal/deployment"
	"github.com/andreyvit/buddyd/internal/director"
	"github.com/andreyvit/buddyd/internal/gracefulshutdown"
)

// -ldflags "-X main.BuildCommit=$(git rev-parse HEAD) -X main.BuildVer=$(git describe --long --dirty)
var (
	BuildCommit string = "unknown"
	BuildVer    string = "unknown"
)

const (
	appName = "Buddy"
)

var (
	openAICreds         openai.Credentials
	baseURLStr          string
	crashOnPanic        bool
	serveAssetsFromDisk bool
	isTesting           bool
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime)

	var (
		env      envloader.VarSet
		required        = envloader.Required
		optional        = envloader.Optional
		bindAddr string = "127.0.0.1"
		bindPort int    = 3003
		dataDir  string
		appOpt   = AppOptions{}
	)
	env.Var("OPENAI_API_KEY", required, envloader.StringVar(&openAICreds.APIKey), "OpenAI API key")
	env.Var("DATA_DIR", required, envloader.StringVar(&dataDir), "path to database directory")
	env.Var("BASE_URL", required, envloader.StringVar(&baseURLStr), "base URL (ex: https://chat.tarantsov.com/)")
	env.Var("BIND", optional, envloader.StringVar(&bindAddr), "network interface to listen on (defaults to 127.0.0.1)")
	env.Var("PORT", optional, envloader.IntVar(&bindPort), "TCP port for HTTP server to listen on")

	var (
		envFile string
	)
	flag.Usage = usage
	flag.Var(env.PrintAction(), "print-env", "print all supported environment variables in shell format")
	flag.Var(action(deployment.RunStdin), "install", "install (aka deploy) this binary using JSON options read from stdin")
	flag.Var(action(func() { fmt.Println(BuildCommit) }), "version", "print version")
	flag.Var(action(func() { fmt.Println(BuildVer) }), "print-commit", "print Git commit ID")
	flag.StringVar(&envFile, "f", ".env", "load environment from this file")
	flag.BoolVar(&crashOnPanic, "crash-on-panic", false, "do not recover from panics to make debugging easier")
	flag.BoolVar(&serveAssetsFromDisk, "assets-from-disk", false, "for speed of development, load static assets from the file system and not from the data embedded into the binary (requires app files to be located at exactly the same path as when it was built, i.e. when building and running on the same machine)")
	flag.Parse()
	loadEnv(envFile)
	env.Parse()

	initializeEmbeddedStatics()

	dir := director.New()
	defer dir.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gracefulshutdown.InterceptShutdownSignals(cancel)

	app := setupApp(dataDir, appOpt)
	defer app.Close()

	ensure(dir.Start(ctx, &director.Component{
		Name:         "http",
		Critical:     true,
		RestartDelay: time.Second,
	}, func(ctx context.Context, quitf func(err error)) error {
		var err error
		_, err = httpserver.Start(ctx, app.webHandler, quitf, httpserver.Options{
			DebugName:               "http",
			Addr:                    bindAddr,
			Port:                    bindPort,
			AcmeEnabled:             false,
			Logf:                    log.Printf,
			GracefulShutdownTimeout: 10 * time.Second,
		})
		log.Printf("%v server listening on %s port %d", appName, bindAddr, bindPort)
		return err
	}))

	dir.Wait()
}

func usage() {
	base := filepath.Base(os.Args[0])
	fmt.Printf("Usage: %s [options]\n\n", base)

	fmt.Printf("Options:\n")
	flag.PrintDefaults()

	fmt.Printf("\nMost options are set via environment variables. Run %s -print-env for a list.\n", base)
}

type action func()

func (_ action) String() string {
	return ""
}

func (_ action) IsBoolFlag() bool {
	return true
}

func (f action) Set(string) error {
	f()
	os.Exit(0)
	return nil
}
