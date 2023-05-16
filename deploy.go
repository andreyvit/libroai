package main

import (
	_ "embed"
	"net/url"
	"os"
	"path/filepath"

	"github.com/andreyvit/buddyd/mvp"
	d "github.com/andreyvit/buddyd/mvp/deploymentutil"
)

func preinstallConfigs(baseSettings *mvp.Settings) {
	settings := fullSettings.From(baseSettings)

	d.InstallDir(settings.Deployment.ServiceDir, 0755, d.Root)
	u := d.NeedUser(settings.Deployment.User)
	d.InstallIfNotExistsFromStdin(settings.KeyringFile, "Keyring file does not exist on the server: {{.path}}.\nAssuming you're initializing a new environment.\nPaste a keyring file, end with an empty line:", 0600, u)
	d.InstallIfNotExistsFromString(settings.LocalOverridesFile, "{}", 0644, d.Root)
}

func install(baseSettings *mvp.Settings) {
	settings := fullSettings.From(baseSettings)

	u := d.NeedUser(settings.Deployment.User)
	serv := settings.Deployment.Service
	mainService := serv + ".service"

	servDir := d.InstallDir(settings.Deployment.ServiceDir, 0755, d.Root)
	binDir := d.InstallDir(filepath.Join(servDir, "bin"), 0755, d.Root)
	dataDir := d.InstallDir(settings.DataDir, 0755, u)

	mainFile := filepath.Join(binDir, serv)

	base := must(url.Parse(settings.BaseURL))

	params := map[string]any{
		"service":  serv,
		"username": settings.Deployment.User,
		"host":     base.Host,
		"port":     settings.BindPort,
		"mainFile": mainFile,
		"servDir":  settings.Deployment.ServiceDir,
		"dataDir":  dataDir,
		// "password": secrets.PasswordCaddy,
		"env": settings.Env,
	}

	d.Install(mainFile, must(os.ReadFile(must(os.Executable()))), 0755, d.Root)
	d.Install("/etc/systemd/system/"+mainService, d.Templ(unitFile, params), 0644, d.Root)
	d.Install(filepath.Join(servDir, "Caddyfile"), d.Templ(caddyfile, params), 0644, d.Root)

	d.Run("systemctl", "daemon-reload")
	d.Run("systemctl", "enable", mainService)
	d.Run("systemctl", "restart", mainService)
	d.Run("/srv/caddy/bin/caddy", "reload", "--config", "/etc/Caddyfile")
}

//	basicauth {
//	    tester {{.password}}
//	}
const caddyfile = `
{{.host}} {
    reverse_proxy :{{.port}}
}
`

const unitFile = `
[Unit]
Description={{.service}}
After=network.target

[Service]
User={{.username}}
Restart=always
PIDFile=/run/{{.service}}.pid
Type=simple
ExecStart={{.mainFile}} -e {{.env}}
KillMode=process

[Install]
WantedBy=multi-user.target
`
