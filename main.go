package main

import (
	"embed"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/andreyvit/envloader"
	"github.com/andreyvit/openai"

	"github.com/andreyvit/buddyd/internal/accesstokens"
	"github.com/andreyvit/buddyd/mvp"
)

// -ldflags "-X main.BuildCommit=$(git rev-parse HEAD) -X main.BuildVer=$(git describe --long --dirty)
var (
	BuildCommit string = "unknown"
	BuildVer    string = "unknown"

	//go:embed config.json
	embeddedConfig []byte
	//go:embed config.secrets.txt
	embeddedSecrets string
	//go:embed views
	embeddedViewsFS embed.FS
	//go:embed static
	embeddedStaticAssetsFS embed.FS
)

var configuration = &mvp.Configuration{
	Envs: map[string][]string{
		"local-andreyvit": {"@all", "@localdevortest", "@localdev"},
		"local-dottedmag": {"@all", "@localdevortest", "@localdev"},
		"stag":            {"@all", "@prodlike"},
		"prod":            {"@all", "@prodlike"},
		"test":            {"@all", "@localdevortest"},
	},

	Preinstall: preinstallConfigs,
	Install:    install,

	BuildCommit: BuildCommit,
	BuildVer:    BuildVer,

	EmbeddedConfig:   embeddedConfig,
	EmbeddedSecrets:  embeddedSecrets,
	EmbeddedStaticFS: embeddedStaticAssetsFS,
	EmbeddedViewsFS:  embeddedViewsFS,

	NewSettings:  newSettings,
	FullSettings: func(base *mvp.Settings) any { return fullSettings(base) },
	NewApp:       newApp,
	NewRC:        newRC,
	LoadSecrets:  loadSecrets,

	Schema: schema,
}

type Settings struct {
	mvp.Settings

	Deployment DeploymentSettings

	OpenAICreds   openai.Credentials
	Password      string
	PasswordCaddy string
}

type DeploymentSettings struct {
	Service    string
	User       string
	ServiceDir string
}

type App struct {
	mvp.App
	users          atomic.Value
	webAdminTokens accesstokens.Configuration
	httpClient     *http.Client
}

type RC struct {
	mvp.RC
}

func main() {
	mvp.Main(configuration)
}

func loadSecrets(baseSettings *mvp.Settings, secr mvp.Secrets) {
	settings := mvp.As[Settings](baseSettings)
	secr.Required("OPENAI_API_KEY", envloader.StringVar(&settings.OpenAICreds.APIKey))
	secr.Required("PASSWORD", envloader.StringVar(&settings.Password))
	secr.Required("PASSWORD_CADDY", envloader.StringVar(&settings.PasswordCaddy))
	secr.Required("POSTMARK_SERVER_TOKEN", envloader.StringVar(&settings.Postmark.ServerAccessToken))
}

func newSettings() *mvp.Settings {
	settings := &Settings{}
	return &settings.Settings
}

func newApp() *mvp.App {
	app := &App{
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}

	app.Hooks.SiteRoutes(mvp.DefaultSite, app.registerRoutes)
	app.Hooks.Helpers(app.registerViewHelpers)

	return &app.App
}

func newRC() *mvp.RC {
	rc := &RC{}
	return &rc.RC
}

func fullSettings(base *mvp.Settings) *Settings {
	return mvp.As[Settings](base)
}
