package mvp

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/andreyvit/buddyd/internal/postmark"
	"github.com/andreyvit/edb"
	"github.com/andreyvit/jsonfix"
	"github.com/andreyvit/plainsecrets"
	"golang.org/x/exp/maps"
)

type Configuration struct {
	Envs map[string][]string

	Preinstall func(settings *Settings)
	Install    func(settings *Settings)

	BuildCommit string
	BuildVer    string

	EmbeddedConfig   []byte
	EmbeddedSecrets  string
	EmbeddedStaticFS embed.FS
	EmbeddedViewsFS  embed.FS

	ConfigFileName  string
	SecretsFileName string
	StaticSubdir    string
	ViewsSubdir     string
	LocalDevAppRoot string

	NewSettings func() *Settings
	FullSettings func(*Settings) any
	NewApp      func() *App
	NewRC       func() *RC
	LoadSecrets func(*Settings, Secrets)

	Schema *edb.Schema
}

func (ge *Configuration) ValidEnvs() []string {
	result := maps.Keys(ge.Envs)
	sort.Strings(result)
	return result
}

type Settings struct {
	Env           string
	Configuration *Configuration

	// Configuration options
	LocalOverridesFile string
	AutoEncryptSecrets bool
	KeyringFile        string

	// HTTP server options
	BindAddr string
	BindPort int

	// job options
	WorkerCount int

	// app options
	AppName string
	BaseURL string
	AppBehaviors

	DataDir string

	Postmark postmark.Credentials
}

type Secrets map[string]string

func (secrets Secrets) Optional(name string, val interface{ Set(string) error }) bool {
	str := secrets[name]
	if str == "" {
		return false
	}
	err := val.Set(str)
	if err != nil {
		log.Fatalf("** ERROR: invalid value of secret %s: %v", name, err)
	}
	return true
}

func (secrets Secrets) Required(name string, val interface{ Set(string) error }) {
	ok := secrets.Optional(name, val)
	if !ok {
		log.Fatalf("** ERROR: missing secret %s", name)
	}
}

func LoadConfig(ge *Configuration, env string, installHook func(*Settings)) *Settings {
	settings := ge.NewSettings()
	settings.Env = env
	settings.Configuration = ge
	full := ge.FullSettings(settings)

	configBySection := make(map[string]json.RawMessage)

	decoder := json.NewDecoder(bytes.NewReader(jsonfix.Bytes(ge.EmbeddedConfig)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&configBySection)
	if err != nil {
		log.Fatalf("** %v", fmt.Errorf("config.json: %w", err))
	}

	for e := range ge.Envs {
		if e == env {
			parseConfigSections(ge, ge.Envs[e], configBySection, full)
		} else {
			// parse all other sections to ensure we're not surprised after deployment
			dummy := ge.NewSettings()
			parseConfigSections(ge, ge.Envs[e], configBySection, ge.FullSettings(dummy))
		}
	}

	if installHook != nil {
		installHook(settings)
	}

	if overridesFile := settings.LocalOverridesFile; overridesFile != "" {
		raw, err := os.ReadFile(overridesFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("** %v", fmt.Errorf("LocalOverridesFile %s is missing, create with {} as its contents", overridesFile))
			}
			log.Fatalf("** %v", fmt.Errorf("%s: %w", overridesFile, err))
		}

		decoder := json.NewDecoder(bytes.NewReader(jsonfix.Bytes(raw)))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(settings)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("%s: %w", overridesFile, err))
		}
	}

	if settings.KeyringFile == "" {
		log.Fatalf("** %v", fmt.Errorf("config.json: empty KeyringFile"))
	}
	keyring, err := plainsecrets.ParseKeyringFile(settings.KeyringFile)
	if err != nil {
		log.Fatalf("** %v", err)
	}

	vals, err := plainsecrets.ParseString(fmt.Sprintf("@all = %s\n%s", strings.Join(maps.Keys(ge.Envs), " "), ge.EmbeddedSecrets))
	if err != nil {
		log.Fatalf("** %v", fmt.Errorf("config.secrets.txt: %w", err))
	}

	if settings.AutoEncryptSecrets {
		_, thisFile, _, _ := runtime.Caller(0)
		if thisFile != "" {
			secretsFile := filepath.Join(filepath.Dir(thisFile), ge.SecretsFileName)
			_, err := os.Stat(secretsFile)
			if err == nil {
				n, failed, err := vals.EncryptAllInFile(secretsFile, keyring)
				if err != nil {
					log.Fatalf("** %v", fmt.Errorf("autoencrypt failed: %w", err))
				}
				if len(failed) > 0 {
					var msgs []string
					var msgSet = make(map[string]bool)
					for _, v := range failed {
						s := v.Err.Error()
						if !msgSet[s] {
							msgSet[s] = true
							msgs = append(msgs, s)
						}
					}
					log.Fatalf("** %v", fmt.Errorf("autoencrypt failed: %s", strings.Join(msgs, ", ")))
				}
				if n > 0 {
					log.Printf("Auto-encrypted %d secret(s).", n)
				}
			}
		}
	}

	secrets, err := vals.EnvValues(env, keyring)
	if err != nil {
		log.Fatalf("** %v", err)
	}
	ge.LoadSecrets(settings, secrets)

	return settings
}

func parseConfigSections(ge *Configuration, sections []string, configBySection map[string]json.RawMessage, settings any) {
	for _, section := range sections {
		if configBySection[section] == nil {
			log.Fatalf("** %v", fmt.Errorf("%s: missing section %s", ge.ConfigFileName, section))
		}
		decoder := json.NewDecoder(bytes.NewReader(configBySection[section]))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(settings)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("%s: %s: %w", ge.ConfigFileName, section, err))
		}
	}
}
