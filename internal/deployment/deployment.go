package deployment

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/exp/slices"
)

type Options struct {
	ServiceName string            `json:"service"`
	Username    string            `json:"user"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	OwnUser     bool              `json:"own_user"`
	Password    string            `json:"password"`
	Env         map[string]string `json:"env"`
}

var (
	//go:embed res/Caddyfile
	caddyfile string
	//go:embed res/unit.service
	unit string
)

func Run(opt Options) {
	log.Printf("Installing with config: %s", must(json.MarshalIndent(opt, "", "  ")))

	u := needUser(opt.Username)
	serv := opt.ServiceName
	mainService := serv + ".service"

	servDir := installDir(filepath.Join("/srv/", serv), 0755, root)
	binDir := installDir(filepath.Join(servDir, "bin"), 0755, root)
	dataDir := installDir(filepath.Join(servDir, "data"), 0755, u)

	confFile := filepath.Join(servDir, serv+".conf")
	mainFile := filepath.Join(binDir, serv)

	params := map[string]any{
		"service":  serv,
		"username": opt.Username,
		"host":     opt.Host,
		"port":     opt.Port,
		"confFile": confFile,
		"mainFile": mainFile,
	}

	updateKEVConfig(confFile, 0644, root, func(conf *KV) {
		conf.Default("VDEMIR_APP_NAME", "Virtual Demir")
		conf.Set("VDEMIR_DATA_DIR", dataDir)
		conf.Set("VDEMIR_PORT_HTTP", strconv.Itoa(opt.Port))
		conf.Set("VDEMIR_URL", "https://"+opt.Host)
		promote(conf, opt.Env, "OPENAI_API_KEY", true, "")
		promote(conf, opt.Env, "OPENAI_ORG", false, "")
		promote(conf, opt.Env, "OPENAI_TUNED_MODEL", false, "none")
		params["password"] = promote(conf, opt.Env, "VDEMIR_PASSWORD", true, "")
		conf.DefaultFunc("VDEMIR_KEYS_PERSISTENT", func() string { return randomHex(256) })
	})

	install(mainFile, must(os.ReadFile(must(os.Executable()))), 0755, root)
	install("/etc/systemd/system/"+mainService, templ(unit, params), 0644, root)
	install(filepath.Join(servDir, "Caddyfile"), templ(caddyfile, params), 0644, root)

	run("systemctl", "daemon-reload")
	run("systemctl", "enable", mainService)
	run("systemctl", "restart", mainService)
	run("/srv/caddy/bin/caddy", "reload", "--config", "/etc/Caddyfile")
}

func RunIfDesired() {
	if i := slices.Index(os.Args, "--install"); i >= 1 {
		RunStdin()
		os.Exit(0)
	}
}

func RunStdin() {
	raw := must(io.ReadAll(os.Stdin))
	var opt Options
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	err := dec.Decode(&opt)
	if err != nil {
		log.Fatalf("install: invalid options: %v", err)
	}
	Run(opt)
}
