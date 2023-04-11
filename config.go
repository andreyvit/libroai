package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/andreyvit/jsonfix"
	"github.com/andreyvit/plainsecrets"
	"golang.org/x/exp/maps"
)

var (
	//go:embed config.json
	configData []byte
	//go:embed config.secrets.txt
	secretsStr string
)

func loadConfig(env string, settings *AppSettings, installHook func()) Secrets {
	configBySection := make(map[string]json.RawMessage)

	decoder := json.NewDecoder(bytes.NewReader(jsonfix.Bytes(configData)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&configBySection)
	if err != nil {
		log.Fatalf("** %v", fmt.Errorf("config.json: %w", err))
	}

	settings.Env = env
	for e := range envs {
		if e == env {
			parseConfigSections(envs[e], configBySection, settings)
		} else {
			// parse all other sections to ensure we're not surprised after deployment
			parseConfigSections(envs[e], configBySection, &AppSettings{})
		}
	}

	if installHook != nil {
		installHook()
	}

	if settings.LocalOverridesFile != "" {
		raw, err := os.ReadFile(settings.LocalOverridesFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("** %v", fmt.Errorf("LocalOverridesFile %s is missing, create with {} as its contents", settings.LocalOverridesFile))
			}
			log.Fatalf("** %v", fmt.Errorf("%s: %w", settings.LocalOverridesFile, err))
		}

		decoder := json.NewDecoder(bytes.NewReader(jsonfix.Bytes(raw)))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(settings)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("%s: %w", settings.LocalOverridesFile, err))
		}
	}

	if settings.KeyringFile == "" {
		log.Fatalf("** %v", fmt.Errorf("config.json: empty KeyringFile"))
	}
	keyring, err := plainsecrets.ParseKeyringFile(settings.KeyringFile)
	if err != nil {
		log.Fatalf("** %v", err)
	}

	vals, err := plainsecrets.ParseString(fmt.Sprintf("@all = %s\n%s", strings.Join(maps.Keys(envs), " "), secretsStr))
	if err != nil {
		log.Fatalf("** %v", fmt.Errorf("config.secrets.txt: %w", err))
	}

	if settings.AutoEncryptSecrets {
		_, thisFile, _, _ := runtime.Caller(0)
		if thisFile != "" {
			secretsFile := filepath.Join(filepath.Dir(thisFile), "config.secrets.txt")
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

	return Secrets(secrets)
}

func parseConfigSections(sections []string, configBySection map[string]json.RawMessage, settings *AppSettings) {
	for _, section := range sections {
		if configBySection[section] == nil {
			log.Fatalf("** %v", fmt.Errorf("config.json: missing section %s", section))
		}
		decoder := json.NewDecoder(bytes.NewReader(configBySection[section]))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(settings)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("config.json: %s: %w", section, err))
		}
	}
}

func autoencryptSecrets(path string, vals *plainsecrets.Values, keyring plainsecrets.Keyring) error {
	n, failed, err := vals.EncryptAllInFile(path, keyring)
	if err != nil {
		return fmt.Errorf("autoencrypt failed: %w", err)
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
		return fmt.Errorf("autoencrypt failed: %s", strings.Join(msgs, ", "))
	}
	if n > 0 {
		log.Printf("Auto-encrypted %d secret(s).", n)
	}
	return nil
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
