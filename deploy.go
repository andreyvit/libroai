package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	d "github.com/andreyvit/buddyd/internal/deployment"
)

func preinstallConfigs() {
	u := d.NeedUser(settings.Deployment.User)

	if !d.Exists(settings.KeyringFile) {
		ensure(os.MkdirAll(filepath.Dir(settings.KeyringFile), 0755))
		log.Printf("Keyring file does not exist on the server: %s", settings.KeyringFile)
		log.Printf("Assuming you're initializing a new environment.")
		log.Printf("Paste a keyring file, end with an empty line:")

		var keyringText bytes.Buffer
		for reader := bufio.NewReader(os.Stdin); ; {
			fmt.Fprintf(os.Stderr, "> ")
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				keyringText.Write(line)
			}
			if err == io.EOF || len(bytes.TrimSpace(line)) == 0 {
				break
			}
			ensure(err)
		}
		d.Install(settings.KeyringFile, keyringText.Bytes(), 0600, u)
	}

	if !d.Exists(settings.LocalOverridesFile) {
		ensure(os.MkdirAll(filepath.Dir(settings.LocalOverridesFile), 0755))
		d.Install(settings.LocalOverridesFile, []byte("{}"), 0644, d.Root)
	}
}

func install() {
	u := d.NeedUser(settings.Deployment.User)
	serv := settings.Deployment.Service
	mainService := serv + ".service"

	servDir := d.InstallDir(settings.Deployment.ServiceDir, 0755, d.Root)
	binDir := d.InstallDir(filepath.Join(servDir, "bin"), 0755, d.Root)
	dataDir := d.InstallDir(settings.Deployment.DataDir, 0755, u)

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
		"password": secrets.PasswordCaddy,
		"env":      settings.Env,
	}

	d.Install(mainFile, must(os.ReadFile(must(os.Executable()))), 0755, d.Root)
	d.Install("/etc/systemd/system/"+mainService, d.Templ(unitFile, params), 0644, d.Root)
	d.Install(filepath.Join(servDir, "Caddyfile"), d.Templ(caddyfile, params), 0644, d.Root)

	d.Run("systemctl", "daemon-reload")
	d.Run("systemctl", "enable", mainService)
	d.Run("systemctl", "restart", mainService)
	d.Run("/srv/caddy/bin/caddy", "reload", "--config", "/etc/Caddyfile")
}

const caddyfile = `
{{.host}} {
    basicauth {
        tester {{.password}}
    }
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
