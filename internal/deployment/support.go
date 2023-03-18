package deployment

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
	"unicode/utf8"
)

type User struct {
	Username string
	Uid      int
	Gid      int
}

var root = &User{Uid: 0, Gid: 0, Username: "root"}

func needUser(username string) *User {
	u, err := user.Lookup(username)
	if err != nil || u == nil {
		log.Fatalf("user %s not found: %v", username, err)
	}
	return &User{Username: u.Username, Uid: atoi(u.Uid), Gid: atoi(u.Gid)}
}

type KV struct {
	Lines    []string
	Modified bool
}

func (kv *KV) Set(key, newValue string) {
	newValue = strings.TrimSpace(newValue)
	replacement := fmt.Sprintf("%s=%s", key, newValue)
	oldValue, i := kv.Find(key)
	if i < 0 {
		kv.Lines = append(kv.Lines, replacement)
		kv.Modified = true
	} else if oldValue != newValue {
		kv.Lines[i] = replacement
		kv.Modified = true
	}
}

func (kv *KV) Default(key, newValue string) {
	_, i := kv.Find(key)
	if i < 0 {
		kv.Lines = append(kv.Lines, fmt.Sprintf("%s=%s", key, strings.TrimSpace(newValue)))
		kv.Modified = true
	}
}

func (kv *KV) DefaultFunc(key string, newValueFunc func() string) {
	_, i := kv.Find(key)
	if i < 0 {
		kv.Lines = append(kv.Lines, fmt.Sprintf("%s=%s", key, strings.TrimSpace(newValueFunc())))
		kv.Modified = true
	}
}

func (kv *KV) Get(key string) string {
	oldValue, _ := kv.Find(key)
	return oldValue
}

func (kv *KV) Exists(key string) bool {
	_, i := kv.Find(key)
	return i >= 0
}

func (kv *KV) Find(key string) (string, int) {
	for i, line := range kv.Lines {
		if strings.HasPrefix(line, key) {
			k, v, ok := strings.Cut(line, "=")
			if ok && strings.TrimSpace(k) == key {
				return strings.TrimSpace(v), i
			}
		}
	}
	return "", -1
}

func promote(conf *KV, env map[string]string, key string, required bool, defaultValue string) string {
	oldValue := conf.Get(key)
	isSet := (oldValue != "")

	v, isProvided := env[key]
	if !isSet && !isProvided && required {
		log.Fatalf("Need initial value of env key %s", key)
	}
	if isProvided {
		conf.Set(key, v)
		return v
	} else if !isSet && defaultValue != "" {
		conf.Set(key, defaultValue)
		return defaultValue
	} else {
		return oldValue
	}
}

// updateKEVConfig modifies config file in Key Equals Value format
func updateKEVConfig(path string, perm os.FileMode, u *User, f func(conf *KV)) bool {
	conf := KV{Lines: loadLines(path)}
	f(&conf)
	if conf.Modified {
		install(path, formatLines(conf.Lines), perm, u)
	}
	return conf.Modified
}

func loadLines(path string) []string {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}
		}
		panic(err)
	}
	lines := strings.Split(string(raw), "\n")
	trimSpaceInLines(lines)
	return trimEmptyLines(lines)
}

func formatLines(lines []string) []byte {
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func trimSpaceInLines(lines []string) {
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
}

func trimEmptyLines(lines []string) []string {
	start, end := 0, len(lines)
	for start < end && lines[start] == "" {
		start++
	}
	for start < end && lines[end-1] == "" {
		end--
	}
	return lines[start:end]
}

func install(path string, content []byte, perm os.FileMode, user *User) {
	log.Printf("▸ %s", path)
	if utf8.Valid(content) {
		indent := strings.Repeat(" ", 12)
		log.Println(indent + strings.TrimSpace(strings.ReplaceAll(string(content), "\n", "\n"+indent)))
		log.Println()
	}
	temp := must(os.CreateTemp(filepath.Dir(path), ".~"+filepath.Base(path)+".*"))
	ensure(temp.Chmod(perm))
	ensure(temp.Chown(user.Uid, user.Gid))
	must(temp.Write(content))
	ensure(temp.Close())
	ensure(os.Rename(temp.Name(), path))
}

func subdir(parent, name string, perm os.FileMode, user *User) string {
	path := filepath.Join(parent, name)
	log.Printf("▸ %s/", path)
	ensureSkippingOSExists(os.Mkdir(path, perm))
	ensure(os.Chown(path, user.Uid, user.Gid))
	return path
}

func installDir(path string, perm os.FileMode, user *User) string {
	log.Printf("▸ %s/", path)
	ensureSkippingOSExists(os.Mkdir(path, perm))
	ensure(os.Chown(path, user.Uid, user.Gid))
	return path
}

func templ(templ string, data map[string]any) []byte {
	t := template.New("")
	_ = must(t.Parse(templ))
	var buf bytes.Buffer
	ensure(t.Execute(&buf, data))
	return buf.Bytes()
}

func run(name string, args ...string) {
	log.Printf("▸ %s", shellQuoteCmdline(name, args...))
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("** %s failed: %v\n%s", name, err, output)
	}
}
