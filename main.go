package main

import (
	"embed"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/envloader"
	"github.com/andreyvit/openai"
	"golang.org/x/time/rate"

	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/jsonext"
	"github.com/andreyvit/mvp/jwt"
	mvpm "github.com/andreyvit/mvp/mvpmodel"

	m "github.com/andreyvit/buddyd/model"
)

// -ldflags "-X main.BuildCommit=$(git rev-parse HEAD) -X main.BuildVer=$(git describe --long --dirty)
var (
	BuildCommit string = "unknown"
	BuildVer    string = "unknown"

	//go:embed config.json
	embeddedConfig []byte
	//go:embed config.secrets.txt
	embeddedSecrets string
	//go:embed all:views
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
	FallbackFormID:   "generic-form",

	LoadSecrets: loadSecrets,

	Modules: []*mvp.Module{
		AppModule,
	},

	Types: map[mvpm.Type][]string{
		mvpm.TypeUser:    {"u"},
		m.TypeWaitlister: {"wl"},
	},
}

type Settings struct {
	mvp.Settings

	Deployment DeploymentSettings

	RootUserEmail string

	OpenAICreds   openai.Credentials
	Password      string
	PasswordCaddy string

	SignInCodeExpiration     jsonext.Duration
	SignInCodeResendInterval jsonext.Duration
}

type DeploymentSettings struct {
	Service    string
	User       string
	ServiceDir string
}

type App struct {
	mvp.App
	users                atomic.Value
	httpClient           *http.Client
	dangerousRateLimiter *rate.Limiter

	runtimeAccountsByID map[m.AccountID]*m.RuntimeAccount
	runtimeAccountsMut  sync.RWMutex
}

func (app *App) Settings() *Settings {
	return fullSettings.From(app.App.Settings)
}

type RC struct {
	mvp.RC
	Session      *m.Session
	User         *m.User
	OriginalUser *m.User
	Account      *m.RuntimeAccount

	Chats   []*m.ChatVM
	Library *m.AccountLibrary
}

func (rc *RC) AccountID() m.AccountID {
	if rc.Account == nil {
		return 0
	}
	return rc.Account.ID
}

func (rc *RC) UserID() m.AccountID {
	if rc.User == nil {
		return 0
	}
	return rc.User.ID
}

func (rc *RC) Check(perm m.Permission, obj mvpm.Object) error {
	return m.CheckAccess(rc.User, perm, rc.AccountID(), obj)
}

func (rc *RC) Can(perm m.Permission, obj mvpm.Object) bool {
	return rc.Check(perm, obj) == nil
}

func main() {
	mvp.Main(configuration)
}

func loadSecrets(baseSettings *mvp.Settings, secr mvp.Secrets) {
	settings := fullSettings.From(baseSettings)
	secr.Required("OPENAI_API_KEY", envloader.StringVar(&settings.OpenAICreds.APIKey))
	secr.Required("PASSWORD", envloader.StringVar(&settings.Password))
	secr.Required("PASSWORD_CADDY", envloader.StringVar(&settings.PasswordCaddy))
	secr.Required("POSTMARK_SERVER_TOKEN", envloader.StringVar(&settings.Postmark.ServerAccessToken))
	secr.RequiredNamedKeySet("AUTH_TOKEN_SECRET", &settings.Configuration.AuthTokenKeys, jwt.MinHS256KeyLen, jwt.MaxHS256KeyLen)
}

func newSettings() *mvp.Settings {
	settings := &Settings{}
	return &settings.Settings
}

func newApp() *App {
	return &App{
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		dangerousRateLimiter: rate.NewLimiter(rate.Every(time.Second*5), 5),
	}
}

func initApp(app *App) {
}

func initDB(app *App, rc *RC) {
	email := app.Settings().RootUserEmail
	if email == "" {
		log.Fatalf("%s: RootUserEmail not configured", app.Settings().Configuration.ConfigFileName)
	}
	acc := ensureAccount(app, rc, "sandbox")
	ensureRootUser(app, rc, email, []m.AccountID{acc.ID})
}

func makeRowKey(app *App, tbl *edb.Table) any {
	if tbl.KeyType() == mvp.FlakeIDType() {
		return app.NewID()
	}
	return nil
}

func ensureRootUser(app *App, rc *RC, email string, accountIDs []m.AccountID) *m.User {
	canon := mvp.CanonicalEmail(email)
	user := edb.Lookup[m.User](rc, UsersByEmail, canon)
	if user == nil {
		name, _, _ := strings.Cut(email, "@")
		user = &m.User{
			ID:        app.NewID(),
			Email:     email,
			EmailNorm: canon,
			Name:      name,
		}
	}
	user.Role = m.UserSystemRoleSuperadmin
	for _, aid := range accountIDs {
		memb := user.Membership(aid)
		if memb == nil {
			memb = &m.UserMembership{
				CreationTime: rc.Now,
				AccountID:    aid,
			}
			user.Memberships = append(user.Memberships, memb)
		}
		memb.Role = m.UserAccountRoleAdmin
		memb.Status = m.UserStatusActive
	}
	edb.Put(rc, user) // nop if unchanged
	return user
}

func ensureAccount(app *App, rc *RC, name string) *m.Account {
	account := edb.Select(edb.FullTableScan[m.Account](rc), func(r *m.Account) bool {
		return r.Name == name
	})
	if account == nil {
		account = &m.Account{
			ID:   app.NewID(),
			Name: name,
		}
		edb.Put(rc, account)
	}
	return account
}

func newRC() *mvp.RC {
	rc := &RC{}
	return &rc.RC
}
