package mvp

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/andreyvit/buddyd/mvp/director"
	"github.com/andreyvit/httpserver"
)

func Main(ge *Configuration) {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	if ge.ConfigFileName == "" {
		ge.ConfigFileName = "config.json"
	}
	if ge.SecretsFileName == "" {
		ge.SecretsFileName = "config-secrets.txt"
	}
	if ge.StaticSubdir == "" {
		ge.StaticSubdir = "static"
	}
	if ge.ViewsSubdir == "" {
		ge.ViewsSubdir = "views"
	}
	if ge.LocalDevAppRoot == "" {
		_, file, _, _ := runtime.Caller(1)
		if file == "" {
			panic("missing source file path in binary")
		}
		ge.LocalDevAppRoot = filepath.Dir(file)
	}
	for k := range ge.Envs {
		ge.Envs[k] = append(ge.Envs[k], k)
	}

	dir := director.New()
	defer dir.Wait()

	var (
		env        string
		installing bool
	)
	flag.Usage = func() {
		base := filepath.Base(os.Args[0])
		fmt.Printf("Usage: %s [options]\n\n", base)

		fmt.Printf("Options:\n")
		flag.PrintDefaults()

		fmt.Printf("\nMost options are set in %s.\n", ge.ConfigFileName)
	}

	flag.StringVar(&env, "e", "", fmt.Sprintf("environment to run, one of %s (defaults to local-$USER)", strings.Join(ge.ValidEnvs(), ", ")))
	flag.BoolVar(&installing, "install", false, "install (aka deploy) this binary")
	flag.Var(action(func() { fmt.Println(ge.BuildCommit) }), "version", "print version")
	flag.Var(action(func() { fmt.Println(ge.BuildVer) }), "print-commit", "print Git commit ID")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	InterceptShutdownSignals(cancel)

	if env == "" && !installing {
		env = "local-" + must(user.Current()).Username
	}

	var installHook func(*Settings)
	if installing {
		installHook = ge.Preinstall
	}

	envGroups := ge.Envs[env]
	if envGroups == nil {
		log.Fatalf("** invalid environment %q, must be one of: %s", env, strings.Join(ge.ValidEnvs(), ", "))
	}
	settings := LoadConfig(ge, env, installHook)

	if installing {
		ge.Install(settings)
		return
	}

	app := ge.NewApp()
	app.Initialize(settings, AppOptions{})
	defer app.Close()

	ensure(dir.Start(ctx, &director.Component{
		Name:         "http",
		Critical:     true,
		RestartDelay: time.Second,
	}, func(ctx context.Context, quitf func(err error)) error {
		var err error
		_, err = httpserver.Start(ctx, app, quitf, httpserver.Options{
			DebugName:               "http",
			Addr:                    settings.BindAddr,
			Port:                    settings.BindPort,
			AcmeEnabled:             false,
			Logf:                    log.Printf,
			GracefulShutdownTimeout: 10 * time.Second,
		})
		log.Printf("%v server listening on %s port %d", settings.AppName, settings.BindAddr, settings.BindPort)
		return err
	}))

	dir.Wait()
}
