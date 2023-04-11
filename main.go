package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andreyvit/envloader"
	"github.com/andreyvit/httpserver"
	"github.com/andreyvit/openai"
	"golang.org/x/exp/maps"

	"github.com/andreyvit/buddyd/internal/director"
	"github.com/andreyvit/buddyd/internal/gracefulshutdown"
)

// -ldflags "-X main.BuildCommit=$(git rev-parse HEAD) -X main.BuildVer=$(git describe --long --dirty)
var (
	BuildCommit string = "unknown"
	BuildVer    string = "unknown"
)

var envs = map[string][]string{
	"local-andreyvit": {"@all", "@local", "@localdev"},
	"local-dottedmag": {"@all", "@local", "@localdev"},
	"stag":            {"@all", "@prodlike"},
	"prod":            {"@all", "@prodlike"},
	"test":            {"@all", "@local"},
}

func init() {
	for k := range envs {
		envs[k] = append(envs[k], k)
	}
}

type AppSettings struct {
	Env                string
	LocalOverridesFile string
	AutoEncryptSecrets bool
	KeyringFile        string

	AppName             string
	BindAddr            string
	BindPort            int
	DBFile              string
	WorkerCount         int
	BaseURL             string
	ServeAssetsFromDisk bool
	CrashOnPanic        bool
	PrettyJSON          bool
	IsTesting           bool
	Deployment          DeploymentSettings
}

type DeploymentSettings struct {
	Service    string
	User       string
	ServiceDir string
	DataDir    string
}

type AppSecrets struct {
	OpenAICreds   openai.Credentials
	Password      string
	PasswordCaddy string
}

var (
	settings AppSettings
	secrets  AppSecrets
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	var (
		env        string
		installing bool
	)
	flag.Usage = usage
	flag.StringVar(&env, "e", "", fmt.Sprintf("environment to run, one of %s (defaults to local-$USER)", strings.Join(validEnvs(), ", ")))
	flag.BoolVar(&installing, "install", false, "install (aka deploy) this binary using JSON options read from stdin")
	flag.Var(action(func() { fmt.Println(BuildCommit) }), "version", "print version")
	flag.Var(action(func() { fmt.Println(BuildVer) }), "print-commit", "print Git commit ID")
	flag.Parse()

	if env == "" && !installing {
		env = "local-" + must(user.Current()).Username
	}
	if installing {
		initApp(env, preinstallConfigs)
		install()
	} else {
		initApp(env, nil)
		runApp()
	}
}

func initApp(env string, installHook func()) {
	envGroups := envs[env]
	if envGroups == nil {
		log.Fatalf("** invalid environment %q, must be one of: %s", env, strings.Join(validEnvs(), ", "))
	}

	secr := loadConfig(env, &settings, installHook)
	secr.Required("OPENAI_API_KEY", envloader.StringVar(&secrets.OpenAICreds.APIKey))
	secr.Required("PASSWORD", envloader.StringVar(&secrets.Password))
	secr.Required("PASSWORD_CADDY", envloader.StringVar(&secrets.PasswordCaddy))

	initializeEmbeddedStatics()
}

func runApp() {
	dir := director.New()
	defer dir.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gracefulshutdown.InterceptShutdownSignals(cancel)

	app := setupApp(settings.DBFile, AppOptions{})
	defer app.Close()

	ensure(dir.Start(ctx, &director.Component{
		Name:         "http",
		Critical:     true,
		RestartDelay: time.Second,
	}, func(ctx context.Context, quitf func(err error)) error {
		var err error
		_, err = httpserver.Start(ctx, app.setupHandler(), quitf, httpserver.Options{
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

func usage() {
	base := filepath.Base(os.Args[0])
	fmt.Printf("Usage: %s [options]\n\n", base)

	fmt.Printf("Options:\n")
	flag.PrintDefaults()

	fmt.Printf("\nMost options are set via environment variables. Run %s -print-env for a list.\n", base)
}

func validEnvs() []string {
	result := maps.Keys(envs)
	sort.Strings(result)
	return result
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
